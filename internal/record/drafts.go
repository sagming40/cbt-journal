package record

// 작성 중 초안 (REQUIREMENTS T3, DATA_MODEL 4.2 · 4.3절). Valkey에 있고, 날아가면 초안만 잃는다 (decisions 0005).
// 비유: 책상 위 연습장이다. 다 쓰면(완료) 공책(Postgres)에 옮겨 적고 연습장은 버린다.

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"

	"github.com/sagming40/cbt-journal/internal/apierr"
)

const (
	draftTTL       = 14 * 24 * time.Hour // DATA_MODEL 4.2절
	maxDrafts      = 10                  // 사용자당. 넘으면 거절한다 — 오래된 걸 몰래 지우지 않는다 (DATA_MODEL 4.3절)
	maxDraftItems  = 30                  // 감정 · 왜곡 칸 하나의 개수. 초안 크기에 끝이 있게
	maxStep        = 8
	previewRunes   = 40 // 이어 쓰기 목록의 상황 앞부분 (DESIGN 4.1절)
	draftOpTimeout = 3 * time.Second
)

// 초안 API만 내는 오류 (ARCHITECTURE 3.2절). 문구는 화면에 그대로 보인다.
var (
	ErrDraftConflict    = errors.New("다른 기기에서 이 초안을 먼저 고쳤어요")
	ErrDraftLimit       = errors.New("초안은 10개까지 둘 수 있어요. 이어 쓰기 목록에서 하나를 지우거나 완료해 주세요")
	ErrDraftUnavailable = errors.New("지금은 초안을 저장할 수 없어요. 쓴 내용은 화면에 그대로 있어요")
)

// DraftEmotion은 초안의 감정 하나다. 강도는 아직 없을 수 있다.
type DraftEmotion struct {
	EmotionID int  `json:"emotion_id"`
	Before    *int `json:"intensity_before"`
	After     *int `json:"intensity_after"`
}

// DraftContent는 화면이 보내는 초안 내용이다. Valkey에는 이 JSON에 id · rev · updated_at을 더해 저장한다 (DATA_MODEL 4.3절).
type DraftContent struct {
	Step               int            `json:"step"`
	OccurredAt         time.Time      `json:"occurred_at"`
	Situation          string         `json:"situation"`
	AutomaticThought   string         `json:"automatic_thought"`
	EvidenceFor        string         `json:"evidence_for"`
	EvidenceAgainst    string         `json:"evidence_against"`
	AlternativeThought string         `json:"alternative_thought"`
	Emotions           []DraftEmotion `json:"emotions"`
	DistortionIDs      []int          `json:"distortion_ids"`
}

// Draft는 저장된 초안 전체다.
type Draft struct {
	DraftContent
	ID        string    `json:"id"`
	Rev       int       `json:"rev"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DraftSummary는 이어 쓰기 목록의 한 줄이다.
type DraftSummary struct {
	ID               string
	Step             int
	SituationPreview string
	UpdatedAt        time.Time
}

// 스크립트 하나가 Valkey 안에서 끊기지 않고 돈다 — "확인 + 저장" 사이에 다른 기기의 요청이 끼어들지 못한다.
// 따로 하면 동시에 온 두 요청이 둘 다 "9개네" 하고 11개가 되거나, 둘 다 rev가 맞다고 보고 한쪽을 덮는다 (DATA_MODEL 4.3절).
// 서버가 새 rev까지 넣어 만든 JSON을 그대로 저장한다: Lua의 cjson으로 다시 만들면 빈 배열 []이 {}로 바뀐다.

// KEYS[1] = 목록. ARGV = 초안 키 앞부분, 제한, 새 id, JSON, TTL(ms), 저장 시각(ms)
// 돌려주는 값: 1 만듦, 0 제한에 걸림.
var createDraftScript = redis.NewScript(`
local ids = redis.call('ZRANGE', KEYS[1], 0, -1)
local alive = 0
for _, id in ipairs(ids) do
  if redis.call('EXISTS', ARGV[1] .. id) == 1 then
    alive = alive + 1
  else
    redis.call('ZREM', KEYS[1], id)
  end
end
if alive >= tonumber(ARGV[2]) then
  return 0
end
redis.call('SET', ARGV[1] .. ARGV[3], ARGV[4], 'PX', ARGV[5])
redis.call('ZADD', KEYS[1], ARGV[6], ARGV[3])
redis.call('PEXPIRE', KEYS[1], ARGV[5])
return 1
`)

// KEYS[1] = 초안, KEYS[2] = 목록. ARGV = 기대 rev, JSON, TTL(ms), 저장 시각(ms), id
// 돌려주는 값: 1 저장함, 0 rev가 다름, -1 초안이 없음.
var saveDraftScript = redis.NewScript(`
local cur = redis.call('GET', KEYS[1])
if not cur then
  redis.call('ZREM', KEYS[2], ARGV[5])
  return -1
end
if cjson.decode(cur).rev ~= tonumber(ARGV[1]) then
  return 0
end
redis.call('SET', KEYS[1], ARGV[2], 'PX', ARGV[3])
redis.call('ZADD', KEYS[2], ARGV[4], ARGV[5])
redis.call('PEXPIRE', KEYS[2], ARGV[3])
return 1
`)

func draftKeyPrefix(userID pgtype.UUID) string { return "draft:" + userID.String() + ":" }
func draftListKey(userID pgtype.UUID) string   { return "drafts:" + userID.String() }

// CreateDraft는 새 초안을 만든다. id는 서버가 정한다.
func (s *Service) CreateDraft(ctx context.Context, userID pgtype.UUID, c DraftContent) (Draft, error) {
	if err := validateDraft(&c); err != nil {
		return Draft{}, err
	}
	if s.rdb == nil {
		return Draft{}, ErrDraftUnavailable
	}
	id, err := newDraftID()
	if err != nil {
		return Draft{}, err
	}
	d := Draft{DraftContent: c, ID: id, Rev: 1, UpdatedAt: s.now().UTC()}
	payload, err := json.Marshal(d)
	if err != nil {
		return Draft{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, draftOpTimeout)
	defer cancel()
	n, err := createDraftScript.Run(ctx, s.rdb, []string{draftListKey(userID)},
		draftKeyPrefix(userID), maxDrafts, id, payload, draftTTL.Milliseconds(), d.UpdatedAt.UnixMilli()).Int()
	if err != nil {
		return Draft{}, s.draftUnavailable(userID, err)
	}
	if n == 0 {
		return Draft{}, ErrDraftLimit
	}
	return d, nil
}

// SaveDraft는 초안을 덮어쓴다. rev가 서버와 같을 때만 저장하고 rev + 1 (DATA_MODEL 4.3절).
func (s *Service) SaveDraft(ctx context.Context, userID pgtype.UUID, id string, rev int, c DraftContent) (Draft, error) {
	uid, ok := parseID(id)
	if !ok {
		return Draft{}, ErrNotFound
	}
	if err := validateDraft(&c); err != nil {
		return Draft{}, err
	}
	if s.rdb == nil {
		return Draft{}, ErrDraftUnavailable
	}
	id = uid.String()
	d := Draft{DraftContent: c, ID: id, Rev: rev + 1, UpdatedAt: s.now().UTC()}
	payload, err := json.Marshal(d)
	if err != nil {
		return Draft{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, draftOpTimeout)
	defer cancel()
	n, err := saveDraftScript.Run(ctx, s.rdb, []string{draftKeyPrefix(userID) + id, draftListKey(userID)},
		rev, payload, draftTTL.Milliseconds(), d.UpdatedAt.UnixMilli(), id).Int()
	if err != nil {
		return Draft{}, s.draftUnavailable(userID, err)
	}
	switch n {
	case -1:
		return Draft{}, ErrNotFound
	case 0:
		return Draft{}, ErrDraftConflict
	}
	return d, nil
}

// GetDraft는 초안 하나를 돌려준다. 키에 user_id가 들어 있어 남의 초안은 없는 것과 같다 (ARCHITECTURE 2절).
func (s *Service) GetDraft(ctx context.Context, userID pgtype.UUID, id string) (Draft, error) {
	uid, ok := parseID(id)
	if !ok {
		return Draft{}, ErrNotFound
	}
	if s.rdb == nil {
		return Draft{}, ErrDraftUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, draftOpTimeout)
	defer cancel()
	raw, err := s.rdb.Get(ctx, draftKeyPrefix(userID)+uid.String()).Bytes()
	if errors.Is(err, redis.Nil) {
		return Draft{}, ErrNotFound
	}
	if err != nil {
		return Draft{}, s.draftUnavailable(userID, err)
	}
	var d Draft
	if err := json.Unmarshal(raw, &d); err != nil {
		return Draft{}, fmt.Errorf("초안 해석: %w", err)
	}
	return d, nil
}

// ListDrafts는 내 초안을 최근 저장순으로 돌려준다. 키가 사라진 id(14일 만료)는 여기서 목록에서 지운다 (DATA_MODEL 4.3절).
func (s *Service) ListDrafts(ctx context.Context, userID pgtype.UUID) ([]DraftSummary, error) {
	if s.rdb == nil {
		return nil, ErrDraftUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, draftOpTimeout)
	defer cancel()
	listKey := draftListKey(userID)
	ids, err := s.rdb.ZRangeArgs(ctx, redis.ZRangeArgs{Key: listKey, Start: 0, Stop: -1, Rev: true}).Result()
	if err != nil {
		return nil, s.draftUnavailable(userID, err)
	}
	out := make([]DraftSummary, 0, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = draftKeyPrefix(userID) + id
	}
	vals, err := s.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, s.draftUnavailable(userID, err)
	}
	var ghosts []any
	for i, v := range vals {
		raw, ok := v.(string)
		if !ok { // nil = 키가 만료됨
			ghosts = append(ghosts, ids[i])
			continue
		}
		var d Draft
		if err := json.Unmarshal([]byte(raw), &d); err != nil {
			return nil, fmt.Errorf("초안 해석: %w", err)
		}
		out = append(out, DraftSummary{ID: d.ID, Step: d.Step, SituationPreview: preview(d.Situation), UpdatedAt: d.UpdatedAt})
	}
	if len(ghosts) > 0 {
		if err := s.rdb.ZRem(ctx, listKey, ghosts...).Err(); err != nil {
			// 돌려줄 목록은 이미 맞다. 다음에 다시 지우면 된다.
			s.logger.Warn("만료된 초안 id를 지우지 못했다", "user_id", userID.String(), "err", err)
		}
	}
	return out, nil
}

// DeleteDraft는 초안을 지운다.
func (s *Service) DeleteDraft(ctx context.Context, userID pgtype.UUID, id string) error {
	uid, ok := parseID(id)
	if !ok {
		return ErrNotFound
	}
	if s.rdb == nil {
		return ErrDraftUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, draftOpTimeout)
	defer cancel()
	n, err := s.deleteDraftKeys(ctx, userID, uid.String())
	if err != nil {
		return s.draftUnavailable(userID, err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// removeDraftAfterComplete는 완료한 초안을 지운다. 기록은 이미 커밋됐으니 실패해도 로그만 남긴다
// — 초안이 남아 있어도 다시 완료하면 있던 기록을 돌려준다 (DATA_MODEL 3.5절).
func (s *Service) removeDraftAfterComplete(ctx context.Context, userID pgtype.UUID, id string) {
	if s.rdb == nil {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, draftOpTimeout)
	defer cancel()
	if _, err := s.deleteDraftKeys(ctx, userID, id); err != nil {
		s.logger.Warn("완료한 초안을 지우지 못했다", "user_id", userID.String(), "draft_id", id, "err", err)
	}
}

// deleteDraftKeys는 초안 키와 목록의 id를 같이 지운다. 돌려주는 값은 지운 초안 키 개수.
func (s *Service) deleteDraftKeys(ctx context.Context, userID pgtype.UUID, id string) (int64, error) {
	pipe := s.rdb.TxPipeline()
	del := pipe.Del(ctx, draftKeyPrefix(userID)+id)
	pipe.ZRem(ctx, draftListKey(userID), id)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return del.Val(), nil
}

// draftUnavailable은 Valkey 오류를 503용 오류로 바꾼다. 내용 없이 id와 원인만 남긴다 (REQUIREMENTS 5.1절).
func (s *Service) draftUnavailable(userID pgtype.UUID, err error) error {
	s.logger.Warn("초안 저장소에 닿지 못했다", "user_id", userID.String(), "err", err)
	return ErrDraftUnavailable
}

// validateDraft는 초안을 느슨하게 검사한다: 빈 칸은 괜찮고 길이 · 범위만 본다. 필수 칸은 완료할 때 본다 (DATA_MODEL 4.3절).
// 글 칸의 앞뒤 공백을 다듬지 않는다 — 쓰는 중인 글이라 이어서 칠 띄어쓰기가 사라지면 안 된다.
func validateDraft(c *DraftContent) error {
	fields := map[string]string{}
	if c.Step < 1 || c.Step > maxStep {
		fields["step"] = "단계는 1에서 8 사이예요"
	}
	if c.OccurredAt.IsZero() {
		fields["occurred_at"] = "일어난 시각이 필요해요"
	}
	long := func(field, value string, limit int, msg string) {
		if utf8.RuneCountInString(value) > limit {
			fields[field] = msg
		}
	}
	long("situation", c.Situation, maxShortText, "1,000자까지 쓸 수 있어요")
	long("automatic_thought", c.AutomaticThought, maxShortText, "1,000자까지 쓸 수 있어요")
	long("alternative_thought", c.AlternativeThought, maxShortText, "1,000자까지 쓸 수 있어요")
	long("evidence_for", c.EvidenceFor, maxLongText, "2,000자까지 쓸 수 있어요")
	long("evidence_against", c.EvidenceAgainst, maxLongText, "2,000자까지 쓸 수 있어요")

	if len(c.Emotions) > maxDraftItems {
		fields["emotions"] = "감정이 너무 많아요"
	}
	seen := map[int]bool{}
	for _, e := range c.Emotions {
		switch {
		case e.EmotionID <= 0:
			fields["emotions"] = "없는 감정이 들어 있어요"
		case seen[e.EmotionID]:
			fields["emotions"] = "같은 감정을 두 번 고를 수 없어요"
		case outOfRange(e.Before) || outOfRange(e.After):
			fields["emotions"] = "강도는 0에서 100 사이예요"
		}
		seen[e.EmotionID] = true
	}
	if len(c.DistortionIDs) > maxDraftItems {
		fields["distortion_ids"] = "왜곡이 너무 많아요"
	}
	seen = map[int]bool{}
	for _, id := range c.DistortionIDs {
		if id <= 0 || seen[id] {
			fields["distortion_ids"] = "왜곡 선택을 다시 확인해 주세요"
		}
		seen[id] = true
	}
	if len(fields) > 0 {
		return &apierr.ValidationError{Fields: fields}
	}
	// 저장하는 JSON에 null 대신 []이 들어가게 (화면이 배열로 믿고 읽는다)
	if c.Emotions == nil {
		c.Emotions = []DraftEmotion{}
	}
	if c.DistortionIDs == nil {
		c.DistortionIDs = []int{}
	}
	c.OccurredAt = c.OccurredAt.UTC()
	return nil
}

func preview(s string) string {
	r := []rune(s)
	if len(r) <= previewRunes {
		return s
	}
	return string(r[:previewRunes])
}

// newDraftID는 무작위 UUID(v4)를 만든다. 초안 id는 추측할 수 없으면 되고 순서가 필요 없다.
func newDraftID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // 버전 4
	b[8] = (b[8] & 0x3f) | 0x80 // 변형 10
	return pgtype.UUID{Bytes: b, Valid: true}.String(), nil
}
