// Package checkin은 빠른 체크인(REQUIREMENTS C1 · C2 · C4)의 규칙과 저장을 맡는다.
// handler(HTTP ↔ 도메인) → service(규칙 · 트랜잭션) → db(sqlc) 한 방향 (ARCHITECTURE 2절).
// 다른 기능 패키지(auth 등)는 import하지 않는다 — 로그인한 사용자 id는 server가 함수로 넘겨준다.
package checkin

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

// 입력 규칙 (DATA_MODEL 3.4절 · DESIGN 4.2절). DB CHECK 제약과 숫자를 맞춘다.
const (
	maxNoteLen   = 200
	maxIntensity = 100
	// 기기 시계가 조금 빨라도 "지금"이 거절되지 않게 앞으로 5분까지는 받는다 (DATA_MODEL 3.4절)
	futureSlack = 5 * time.Minute
)

// ErrNotFound: 없는 체크인. 남의 것도 같은 오류다 — 있는지조차 알려 주지 않는다 (ARCHITECTURE 2절).
var ErrNotFound = errors.New("찾을 수 없는 기록이에요")

// Emotion은 고를 수 있는 감정 하나다.
type Emotion struct {
	ID   int16
	Name string
}

// EmotionEntry는 체크인에 고른 감정 하나와 그 강도다.
type EmotionEntry struct {
	EmotionID int16
	Name      string
	Intensity int16
}

// Checkin은 화면에 돌려줄 체크인 하나다.
type Checkin struct {
	ID         pgtype.UUID
	OccurredAt time.Time
	Note       string
	Emotions   []EmotionEntry // 감정 화면 순서대로
}

// InputEmotion은 고른 감정 하나(입력).
type InputEmotion struct {
	EmotionID int
	Intensity int
}

// Input은 만들기 · 고치기 입력이다. OccurredAt이 nil이면 지금.
type Input struct {
	OccurredAt *time.Time
	Note       string
	Emotions   []InputEmotion
}

// Service는 체크인 규칙을 한곳에 모은다.
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
	Valkey *redis.Client // 통계 캐시 버전만 올린다. 죽어 있어도 저장은 된다 (decisions 0005)
	Logger *slog.Logger
	Now    func() time.Time // nil이면 time.Now (테스트에서 시계를 고정할 때)
}

// NewService는 체크인 서비스를 만든다.
func NewService(o Options) *Service {
	now := o.Now
	if now == nil {
		now = time.Now
	}
	return &Service{pool: o.Pool, q: dbgen.New(o.Pool), rdb: o.Valkey, logger: o.Logger, now: now}
}

// Emotions는 고를 수 있는 감정을 화면 순서대로 돌려준다.
func (s *Service) Emotions(ctx context.Context) ([]Emotion, error) {
	rows, err := s.q.ListEmotions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Emotion, len(rows))
	for i, r := range rows {
		out[i] = Emotion{ID: r.ID, Name: r.Name}
	}
	return out, nil
}

// Day는 그날(사용자 시간대 기준)의 체크인을 시간순으로 돌려준다. date가 ""이면 오늘.
// 돌려주는 날짜는 실제로 쓴 날짜(YYYY-MM-DD)다 — 화면이 "오늘"을 서버와 같은 기준으로 보이게.
func (s *Service) Day(ctx context.Context, userID pgtype.UUID, date string) (string, []Checkin, error) {
	tz, err := s.q.GetUserTimezone(ctx, userID)
	if err != nil {
		return "", nil, err
	}
	day, from, to, err := localday.Range(date, tz, s.now())
	if err != nil {
		return "", nil, err
	}
	rows, err := s.q.ListCheckinsBetween(ctx, dbgen.ListCheckinsBetweenParams{
		UserID: userID, FromTime: ts(from), ToTime: ts(to),
	})
	if err != nil {
		return "", nil, err
	}
	items, err := s.withEmotions(ctx, userID, rows)
	return day, items, err
}

// Get은 체크인 하나를 돌려준다(수정 화면용).
func (s *Service) Get(ctx context.Context, userID pgtype.UUID, id string) (Checkin, error) {
	cid, ok := parseID(id)
	if !ok {
		return Checkin{}, ErrNotFound
	}
	row, err := s.q.GetCheckin(ctx, dbgen.GetCheckinParams{ID: cid, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return Checkin{}, ErrNotFound
	}
	if err != nil {
		return Checkin{}, err
	}
	return s.one(ctx, userID, row)
}

// Create는 체크인을 남긴다. 체크인 행과 감정 행을 한 트랜잭션으로 — 감정 없는 체크인이 남지 않게 (DATA_MODEL 3.4절).
func (s *Service) Create(ctx context.Context, userID pgtype.UUID, in Input) (Checkin, error) {
	v, err := s.validate(ctx, in)
	if err != nil {
		return Checkin{}, err
	}
	var row dbgen.Checkin
	err = s.inTx(ctx, func(q *dbgen.Queries) error {
		var err error
		row, err = q.CreateCheckin(ctx, dbgen.CreateCheckinParams{UserID: userID, OccurredAt: ts(v.occurredAt), Note: v.note})
		if err != nil {
			return err
		}
		return q.AddCheckinEmotions(ctx, dbgen.AddCheckinEmotionsParams{CheckinID: row.ID, EmotionIds: v.emotionIDs, Intensities: v.intensities})
	})
	if err != nil {
		return Checkin{}, err
	}
	s.bumpStats(ctx, userID)
	return s.one(ctx, userID, row)
}

// Update는 체크인을 통째로 고친다(C4). 나중에 저장한 쪽이 이긴다 (DATA_MODEL 3.4절).
func (s *Service) Update(ctx context.Context, userID pgtype.UUID, id string, in Input) (Checkin, error) {
	cid, ok := parseID(id)
	if !ok {
		return Checkin{}, ErrNotFound
	}
	v, err := s.validate(ctx, in)
	if err != nil {
		return Checkin{}, err
	}
	var row dbgen.Checkin
	err = s.inTx(ctx, func(q *dbgen.Queries) error {
		var err error
		// WHERE id AND user_id: 남의 체크인은 0행 → 404. 이 줄이 주인 확인이라, 아래 감정 지우기는 반드시 이 뒤에.
		row, err = q.UpdateCheckin(ctx, dbgen.UpdateCheckinParams{OccurredAt: ts(v.occurredAt), Note: v.note, ID: cid, UserID: userID})
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		if err := q.DeleteCheckinEmotions(ctx, row.ID); err != nil {
			return err
		}
		return q.AddCheckinEmotions(ctx, dbgen.AddCheckinEmotionsParams{CheckinID: row.ID, EmotionIds: v.emotionIDs, Intensities: v.intensities})
	})
	if err != nil {
		return Checkin{}, err
	}
	s.bumpStats(ctx, userID)
	return s.one(ctx, userID, row)
}

// Delete는 체크인을 지운다. 감정 행은 CASCADE로 같이 (DATA_MODEL 3.9절).
func (s *Service) Delete(ctx context.Context, userID pgtype.UUID, id string) error {
	cid, ok := parseID(id)
	if !ok {
		return ErrNotFound
	}
	n, err := s.q.DeleteCheckin(ctx, dbgen.DeleteCheckinParams{ID: cid, UserID: userID})
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
	occurredAt  time.Time
	note        string
	emotionIDs  []int16
	intensities []int16
}

// validate는 입력을 검사한다. 틀린 칸을 한꺼번에 모아 400으로 돌려준다(하나씩 고치며 여러 번 보내지 않게).
// 감정 이름은 숨기지 않은 기본 감정만 받는다.
func (s *Service) validate(ctx context.Context, in Input) (valid, error) {
	fields := map[string]string{}
	v := valid{note: strings.TrimSpace(in.Note)}

	if utf8.RuneCountInString(v.note) > maxNoteLen {
		fields["note"] = "메모는 200자까지 쓸 수 있어요"
	}

	now := s.now()
	v.occurredAt = now
	if in.OccurredAt != nil {
		v.occurredAt = *in.OccurredAt
		if v.occurredAt.After(now.Add(futureSlack)) {
			fields["occurred_at"] = "아직 오지 않은 시각은 고를 수 없어요"
		}
	}

	emotions, err := s.Emotions(ctx)
	if err != nil {
		return valid{}, err
	}
	known := make(map[int]bool, len(emotions))
	for _, e := range emotions {
		known[int(e.ID)] = true
	}
	seen := map[int]bool{}
	switch {
	case len(in.Emotions) == 0:
		fields["emotions"] = "감정을 하나 이상 골라 주세요"
	default:
		for _, e := range in.Emotions {
			switch {
			case !known[e.EmotionID]:
				fields["emotions"] = "없는 감정이 들어 있어요"
			case seen[e.EmotionID]:
				fields["emotions"] = "같은 감정을 두 번 고를 수 없어요"
			case e.Intensity < 0 || e.Intensity > maxIntensity:
				fields["emotions"] = "강도는 0에서 100 사이예요"
			}
			seen[e.EmotionID] = true
			// int16으로 바꾸는 건 검사를 통과한 값만 저장에 쓰이므로 안전하다(범위는 위에서 확인).
			v.emotionIDs = append(v.emotionIDs, int16(e.EmotionID))   //nolint:gosec // known에 있는 id = smallint
			v.intensities = append(v.intensities, int16(e.Intensity)) //nolint:gosec // 0 ~ 100 확인함
		}
	}

	if len(fields) > 0 {
		return valid{}, &apierr.ValidationError{Fields: fields}
	}
	return v, nil
}

// one은 체크인 한 행에 감정을 붙인다.
func (s *Service) one(ctx context.Context, userID pgtype.UUID, row dbgen.Checkin) (Checkin, error) {
	items, err := s.withEmotions(ctx, userID, []dbgen.Checkin{row})
	if err != nil {
		return Checkin{}, err
	}
	return items[0], nil
}

// withEmotions는 여러 체크인의 감정을 쿼리 한 번으로 붙인다(체크인마다 따로 부르지 않게).
func (s *Service) withEmotions(ctx context.Context, userID pgtype.UUID, rows []dbgen.Checkin) ([]Checkin, error) {
	out := make([]Checkin, len(rows))
	if len(rows) == 0 {
		return out, nil
	}
	ids := make([]pgtype.UUID, len(rows))
	index := make(map[[16]byte]int, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
		index[r.ID.Bytes] = i
		out[i] = Checkin{ID: r.ID, OccurredAt: r.OccurredAt.Time, Note: r.Note, Emotions: []EmotionEntry{}}
	}
	ems, err := s.q.ListCheckinEmotions(ctx, dbgen.ListCheckinEmotionsParams{UserID: userID, CheckinIds: ids})
	if err != nil {
		return nil, err
	}
	for _, e := range ems {
		i := index[e.CheckinID.Bytes]
		out[i].Emotions = append(out[i].Emotions, EmotionEntry{EmotionID: e.EmotionID, Name: e.Name, Intensity: e.Intensity})
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

// bumpStats는 통계 캐시 버전을 올린다 — 옛 집계 키는 아무도 안 읽다가 TTL로 사라진다 (DATA_MODEL 4.4절).
// Valkey가 죽어 있어도 저장은 이미 끝났으니 로그만 남긴다(기록 내용은 남기지 않는다).
func (s *Service) bumpStats(ctx context.Context, userID pgtype.UUID) {
	if s.rdb == nil {
		return
	}
	if err := cache.BumpStatsVersion(ctx, s.rdb, userID.String()); err != nil {
		s.logger.Warn("통계 캐시 버전을 올리지 못했다", "user_id", userID.String(), "err", err)
	}
}

// parseID는 주소의 체크인 id를 읽는다. 모양이 틀리면 false → 404(없는 것과 같은 응답).
func parseID(id string) (pgtype.UUID, bool) {
	var u pgtype.UUID
	if err := u.Scan(id); err != nil || !u.Valid {
		return pgtype.UUID{}, false
	}
	return u, true
}

func ts(t time.Time) pgtype.Timestamptz { return pgtype.Timestamptz{Time: t, Valid: true} }
