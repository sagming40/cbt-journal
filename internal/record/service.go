// Package record는 사고 기록(REQUIREMENTS T1 · T5 · T6)과 작성 중 초안(T3)을 맡는다.
// handler(HTTP ↔ 도메인) → service(규칙 · 트랜잭션) → db(sqlc) 한 방향 (ARCHITECTURE 2절).
// 완료한 기록은 Postgres(이 파일), 작성 중 초안은 Valkey(drafts.go)에 있다 (DATA_MODEL 3.5 · 4.3절).
package record

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/sagming40/cbt-journal/internal/apierr"
	"github.com/sagming40/cbt-journal/internal/cache"
	"github.com/sagming40/cbt-journal/internal/db/dbgen"
	"github.com/sagming40/cbt-journal/internal/localday"
)

// 입력 규칙 (DATA_MODEL 3.5절). DB CHECK 제약과 숫자를 맞춘다.
const (
	maxShortText = 1000 // 상황 · 자동 사고 · 대안 사고
	maxLongText  = 2000 // 근거 · 반대 근거
	maxIntensity = 100
	// 기기 시계가 조금 빨라도 "지금"이 거절되지 않게 앞으로 5분까지는 받는다 (DATA_MODEL 3.4절)
	futureSlack = 5 * time.Minute
)

// ErrNotFound: 없는 기록 · 초안. 남의 것도 같은 오류다 — 있는지조차 알려 주지 않는다 (ARCHITECTURE 2절).
var ErrNotFound = errors.New("찾을 수 없는 기록이에요")

// Distortion은 고를 수 있는 인지 왜곡 하나다. 설명 · 예시는 사용자가 쓰기 전에는 nil (DATA_MODEL 3.3절).
type Distortion struct {
	ID          int16
	Name        string
	Description *string
	Example     *string
}

// EmotionEntry는 기록에 고른 감정 하나와 전 · 후 강도다.
type EmotionEntry struct {
	EmotionID int16
	Name      string
	Before    int16
	After     int16
}

// DistortionEntry는 기록에 고른 왜곡 하나다.
type DistortionEntry struct {
	ID   int16
	Name string
}

// Record는 화면에 돌려줄 사고 기록 하나다.
type Record struct {
	ID                 pgtype.UUID
	OccurredAt         time.Time
	Situation          string
	AutomaticThought   string
	EvidenceFor        string
	EvidenceAgainst    string
	AlternativeThought string
	Emotions           []EmotionEntry    // 감정 화면 순서대로
	Distortions        []DistortionEntry // 왜곡 화면 순서대로
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// InputEmotion은 고른 감정 하나(입력). 강도가 nil이면 아직 안 매긴 것 — 완료 · 수정에서는 거절한다.
type InputEmotion struct {
	EmotionID int
	Before    *int
	After     *int
}

// Input은 완료 · 수정 입력이다. OccurredAt이 nil이면 지금.
type Input struct {
	OccurredAt         *time.Time
	Situation          string
	AutomaticThought   string
	EvidenceFor        string
	EvidenceAgainst    string
	AlternativeThought string
	Emotions           []InputEmotion
	DistortionIDs      []int
}

// Service는 사고 기록 규칙을 한곳에 모은다.
type Service struct {
	pool   *pgxpool.Pool
	q      *dbgen.Queries
	rdb    *redis.Client
	logger *slog.Logger
	now    func() time.Time
}

// Options는 Service를 만들 때 넘기는 값이다.
type Options struct {
	Pool   *pgxpool.Pool
	Valkey *redis.Client // 초안 저장소 + 통계 캐시 버전. 죽어 있으면 초안 API만 503이고 완료는 된다 (DATA_MODEL 4.3절)
	Logger *slog.Logger
	Now    func() time.Time // nil이면 time.Now (테스트에서 시계를 고정할 때)
}

// NewService는 사고 기록 서비스를 만든다.
func NewService(o Options) *Service {
	now := o.Now
	if now == nil {
		now = time.Now
	}
	return &Service{pool: o.Pool, q: dbgen.New(o.Pool), rdb: o.Valkey, logger: o.Logger, now: now}
}

// Distortions는 고를 수 있는 인지 왜곡을 화면 순서대로 돌려준다.
func (s *Service) Distortions(ctx context.Context) ([]Distortion, error) {
	rows, err := s.q.ListDistortions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Distortion, len(rows))
	for i, r := range rows {
		out[i] = Distortion{ID: r.ID, Name: r.Name, Description: textPtr(r.Description), Example: textPtr(r.Example)}
	}
	return out, nil
}

// Day는 그날(사용자 시간대 기준)의 사고 기록을 시간순으로 돌려준다. date가 ""이면 오늘.
func (s *Service) Day(ctx context.Context, userID pgtype.UUID, date string) (string, []Record, error) {
	tz, err := s.q.GetUserTimezone(ctx, userID)
	if err != nil {
		return "", nil, err
	}
	day, from, to, err := localday.Range(date, tz, s.now())
	if err != nil {
		return "", nil, err
	}
	rows, err := s.q.ListRecordsBetween(ctx, dbgen.ListRecordsBetweenParams{
		UserID: userID, FromTime: ts(from), ToTime: ts(to),
	})
	if err != nil {
		return "", nil, err
	}
	items, err := s.withChildren(ctx, userID, rows)
	return day, items, err
}

// Get은 사고 기록 하나를 돌려준다.
func (s *Service) Get(ctx context.Context, userID pgtype.UUID, id string) (Record, error) {
	rid, ok := parseID(id)
	if !ok {
		return Record{}, ErrNotFound
	}
	row, err := s.q.GetRecord(ctx, dbgen.GetRecordParams{ID: rid, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrNotFound
	}
	if err != nil {
		return Record{}, err
	}
	return s.one(ctx, userID, row)
}

// Create는 사고 기록을 완료한다. 본문 · 감정 · 왜곡을 한 트랜잭션으로 넣고, 커밋이 성공한 뒤에 초안을 지운다
// — 실패하면 초안이 남아 다시 시도할 수 있다 (DATA_MODEL 4.3절). 내용은 화면이 보낸 본문이 전부다(Valkey를 읽지 않는다).
// 같은 초안으로 다시 오면 새로 만들지 않고 있던 기록을 돌려준다(created = false, DATA_MODEL 3.5절).
func (s *Service) Create(ctx context.Context, userID pgtype.UUID, in Input, draftID string) (rec Record, created bool, err error) {
	var source pgtype.UUID
	if draftID != "" {
		var ok bool
		if source, ok = parseID(draftID); !ok {
			return Record{}, false, apierr.Field("draft_id", "초안 번호가 맞지 않아요")
		}
	}
	v, err := s.validate(ctx, in)
	if err != nil {
		return Record{}, false, err
	}
	var row dbgen.ThoughtRecord
	err = s.inTx(ctx, func(q *dbgen.Queries) error {
		var err error
		row, err = q.CreateRecord(ctx, dbgen.CreateRecordParams{
			UserID: userID, OccurredAt: ts(v.occurredAt), Situation: v.situation, AutomaticThought: v.automatic,
			EvidenceFor: v.evidenceFor, EvidenceAgainst: v.evidenceAgainst, AlternativeThought: v.alternative,
			SourceDraftID: source,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			// ON CONFLICT DO NOTHING: 이미 이 초안으로 완료했다. 자식 행은 그때 넣었다.
			row, err = q.GetRecordByDraft(ctx, dbgen.GetRecordByDraftParams{UserID: userID, SourceDraftID: source})
			return err
		}
		if err != nil {
			return err
		}
		created = true
		return addChildren(ctx, q, row.ID, v)
	})
	if err != nil {
		return Record{}, false, err
	}
	if source.Valid {
		// 다시 온 요청이어도 지운다 — 앞선 요청이 커밋 뒤 지우기에 실패했을 수 있다.
		s.removeDraftAfterComplete(ctx, userID, source.String())
	}
	if created {
		s.bumpStats(ctx, userID)
	}
	rec, err = s.one(ctx, userID, row)
	return rec, created, err
}

// Update는 사고 기록을 통째로 고친다(T5). 나중에 저장한 쪽이 이긴다 (DATA_MODEL 3.5절).
func (s *Service) Update(ctx context.Context, userID pgtype.UUID, id string, in Input) (Record, error) {
	rid, ok := parseID(id)
	if !ok {
		return Record{}, ErrNotFound
	}
	v, err := s.validate(ctx, in)
	if err != nil {
		return Record{}, err
	}
	var row dbgen.ThoughtRecord
	err = s.inTx(ctx, func(q *dbgen.Queries) error {
		var err error
		// WHERE id AND user_id: 남의 기록은 0행 → 404. 이 줄이 주인 확인이라, 아래 자식 지우기는 반드시 이 뒤에.
		row, err = q.UpdateRecord(ctx, dbgen.UpdateRecordParams{
			ID: rid, UserID: userID, OccurredAt: ts(v.occurredAt), Situation: v.situation, AutomaticThought: v.automatic,
			EvidenceFor: v.evidenceFor, EvidenceAgainst: v.evidenceAgainst, AlternativeThought: v.alternative,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if err := q.DeleteRecordEmotions(ctx, row.ID); err != nil {
			return err
		}
		if err := q.DeleteRecordDistortions(ctx, row.ID); err != nil {
			return err
		}
		return addChildren(ctx, q, row.ID, v)
	})
	if err != nil {
		return Record{}, err
	}
	s.bumpStats(ctx, userID)
	return s.one(ctx, userID, row)
}

// Delete는 사고 기록을 지운다. 감정 · 왜곡 행은 CASCADE로 같이 (DATA_MODEL 3.9절).
func (s *Service) Delete(ctx context.Context, userID pgtype.UUID, id string) error {
	rid, ok := parseID(id)
	if !ok {
		return ErrNotFound
	}
	n, err := s.q.DeleteRecord(ctx, dbgen.DeleteRecordParams{ID: rid, UserID: userID})
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	s.bumpStats(ctx, userID)
	return nil
}

// valid는 검사를 통과해 저장할 모양으로 다듬은 입력이다.
type valid struct {
	occurredAt      time.Time
	situation       string
	automatic       string
	evidenceFor     string
	evidenceAgainst string
	alternative     string
	emotionIDs      []int16
	befores         []int16
	afters          []int16
	distortionIDs   []int16
}

// validate는 완료 · 수정 입력을 엄격하게 검사한다(초안의 느슨한 검사는 drafts.go).
// 틀린 칸을 한꺼번에 모아 400으로 돌려준다 — 화면은 그 칸이 있는 가장 앞 단계로 간다 (DESIGN 4.3절).
func (s *Service) validate(ctx context.Context, in Input) (valid, error) {
	fields := map[string]string{}
	v := valid{
		situation:       strings.TrimSpace(in.Situation),
		automatic:       strings.TrimSpace(in.AutomaticThought),
		evidenceFor:     strings.TrimSpace(in.EvidenceFor),
		evidenceAgainst: strings.TrimSpace(in.EvidenceAgainst),
		alternative:     strings.TrimSpace(in.AlternativeThought),
	}
	required := func(field, value, emptyMsg string) {
		switch {
		case value == "":
			fields[field] = emptyMsg
		case utf8.RuneCountInString(value) > maxShortText:
			fields[field] = "1,000자까지 쓸 수 있어요"
		}
	}
	required("situation", v.situation, "어떤 상황이었는지 적어 주세요")
	required("automatic_thought", v.automatic, "떠오른 생각을 적어 주세요")
	required("alternative_thought", v.alternative, "다르게 본 생각을 적어 주세요")
	if utf8.RuneCountInString(v.evidenceFor) > maxLongText {
		fields["evidence_for"] = "2,000자까지 쓸 수 있어요"
	}
	if utf8.RuneCountInString(v.evidenceAgainst) > maxLongText {
		fields["evidence_against"] = "2,000자까지 쓸 수 있어요"
	}

	now := s.now()
	v.occurredAt = now
	if in.OccurredAt != nil {
		v.occurredAt = *in.OccurredAt
		if v.occurredAt.After(now.Add(futureSlack)) {
			fields["occurred_at"] = "아직 오지 않은 시각은 고를 수 없어요"
		}
	}

	if err := s.validateEmotions(ctx, in.Emotions, &v, fields); err != nil {
		return valid{}, err
	}
	if err := s.validateDistortions(ctx, in.DistortionIDs, &v, fields); err != nil {
		return valid{}, err
	}

	if len(fields) > 0 {
		return valid{}, &apierr.ValidationError{Fields: fields}
	}
	return v, nil
}

func (s *Service) validateEmotions(ctx context.Context, in []InputEmotion, v *valid, fields map[string]string) error {
	if len(in) == 0 {
		fields["emotions"] = "감정을 하나 이상 골라 주세요"
		return nil
	}
	// 숨기지 않은 기본 감정만 받는다.
	rows, err := s.q.ListEmotions(ctx)
	if err != nil {
		return err
	}
	known := make(map[int]bool, len(rows))
	for _, r := range rows {
		known[int(r.ID)] = true
	}
	seen := map[int]bool{}
	for _, e := range in {
		msg := ""
		switch {
		case !known[e.EmotionID]:
			msg = "없는 감정이 들어 있어요"
		case seen[e.EmotionID]:
			msg = "같은 감정을 두 번 고를 수 없어요"
		case e.Before == nil || e.After == nil:
			// 고른 감정마다 전 · 후를 전부 매긴다 (DATA_MODEL 3.5절)
			msg = "감정마다 전 · 후 강도를 매겨 주세요"
		case outOfRange(e.Before) || outOfRange(e.After):
			msg = "강도는 0에서 100 사이예요"
		}
		seen[e.EmotionID] = true
		if msg != "" {
			fields["emotions"] = msg
			continue
		}
		v.emotionIDs = append(v.emotionIDs, int16(e.EmotionID)) //nolint:gosec // known에 있는 id = smallint
		v.befores = append(v.befores, int16(*e.Before))         //nolint:gosec // 0 ~ 100 확인함
		v.afters = append(v.afters, int16(*e.After))            //nolint:gosec // 0 ~ 100 확인함
	}
	return nil
}

func (s *Service) validateDistortions(ctx context.Context, in []int, v *valid, fields map[string]string) error {
	v.distortionIDs = make([]int16, 0, len(in))
	if len(in) == 0 {
		return nil // 왜곡을 못 찾는 날도 기록은 남긴다 (DATA_MODEL 3.5절)
	}
	rows, err := s.q.ListDistortions(ctx)
	if err != nil {
		return err
	}
	known := make(map[int]bool, len(rows))
	for _, r := range rows {
		known[int(r.ID)] = true
	}
	seen := map[int]bool{}
	for _, id := range in {
		switch {
		case !known[id]:
			fields["distortion_ids"] = "없는 왜곡이 들어 있어요"
			continue
		case seen[id]:
			fields["distortion_ids"] = "같은 왜곡을 두 번 고를 수 없어요"
			continue
		}
		seen[id] = true
		v.distortionIDs = append(v.distortionIDs, int16(id)) //nolint:gosec // known에 있는 id = smallint
	}
	return nil
}

func outOfRange(n *int) bool { return n != nil && (*n < 0 || *n > maxIntensity) }

// addChildren은 감정 · 왜곡 행을 넣는다. 본문과 같은 트랜잭션에서만 부른다.
func addChildren(ctx context.Context, q *dbgen.Queries, id pgtype.UUID, v valid) error {
	if err := q.AddRecordEmotions(ctx, dbgen.AddRecordEmotionsParams{
		RecordID: id, EmotionIds: v.emotionIDs, Befores: v.befores, Afters: v.afters,
	}); err != nil {
		return err
	}
	return q.AddRecordDistortions(ctx, dbgen.AddRecordDistortionsParams{RecordID: id, DistortionIds: v.distortionIDs})
}

// one은 기록 한 행에 감정 · 왜곡을 붙인다.
func (s *Service) one(ctx context.Context, userID pgtype.UUID, row dbgen.ThoughtRecord) (Record, error) {
	items, err := s.withChildren(ctx, userID, []dbgen.ThoughtRecord{row})
	if err != nil {
		return Record{}, err
	}
	return items[0], nil
}

// withChildren은 여러 기록의 감정 · 왜곡을 쿼리 두 번으로 붙인다(기록마다 따로 부르지 않게).
func (s *Service) withChildren(ctx context.Context, userID pgtype.UUID, rows []dbgen.ThoughtRecord) ([]Record, error) {
	out := make([]Record, len(rows))
	if len(rows) == 0 {
		return out, nil
	}
	ids := make([]pgtype.UUID, len(rows))
	index := make(map[[16]byte]int, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
		index[r.ID.Bytes] = i
		out[i] = Record{
			ID: r.ID, OccurredAt: r.OccurredAt.Time, Situation: r.Situation, AutomaticThought: r.AutomaticThought,
			EvidenceFor: r.EvidenceFor, EvidenceAgainst: r.EvidenceAgainst, AlternativeThought: r.AlternativeThought,
			CreatedAt: r.CreatedAt.Time, UpdatedAt: r.UpdatedAt.Time,
			Emotions: []EmotionEntry{}, Distortions: []DistortionEntry{},
		}
	}
	ems, err := s.q.ListRecordEmotions(ctx, dbgen.ListRecordEmotionsParams{UserID: userID, RecordIds: ids})
	if err != nil {
		return nil, err
	}
	for _, e := range ems {
		i := index[e.RecordID.Bytes]
		out[i].Emotions = append(out[i].Emotions, EmotionEntry{EmotionID: e.EmotionID, Name: e.Name, Before: e.IntensityBefore, After: e.IntensityAfter})
	}
	ds, err := s.q.ListRecordDistortions(ctx, dbgen.ListRecordDistortionsParams{UserID: userID, RecordIds: ids})
	if err != nil {
		return nil, err
	}
	for _, d := range ds {
		i := index[d.RecordID.Bytes]
		out[i].Distortions = append(out[i].Distortions, DistortionEntry{ID: d.DistortionID, Name: d.Name})
	}
	return out, nil
}

// inTx는 fn을 한 트랜잭션 안에서 돌린다. fn이 오류를 내면 되돌린다.
func (s *Service) inTx(ctx context.Context, fn func(q *dbgen.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	// Commit이 성공한 뒤의 Rollback은 아무 일도 하지 않는다 — 실패 경로를 빠뜨리지 않으려고 미리 걸어 둔다.
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(s.q.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// bumpStats는 통계 캐시 버전을 올린다 (DATA_MODEL 4.4절). Valkey가 죽어 있어도 저장은 이미 끝났으니 로그만 남긴다.
func (s *Service) bumpStats(ctx context.Context, userID pgtype.UUID) {
	if s.rdb == nil {
		return
	}
	if err := cache.BumpStatsVersion(ctx, s.rdb, userID.String()); err != nil {
		s.logger.Warn("통계 캐시 버전을 올리지 못했다", "user_id", userID.String(), "err", err)
	}
}

// parseID는 주소의 id를 읽는다. 모양이 틀리면 false → 404(없는 것과 같은 응답).
func parseID(id string) (pgtype.UUID, bool) {
	var u pgtype.UUID
	if err := u.Scan(id); err != nil || !u.Valid {
		return pgtype.UUID{}, false
	}
	return u, true
}

func textPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func ts(t time.Time) pgtype.Timestamptz { return pgtype.Timestamptz{Time: t, Valid: true} }
