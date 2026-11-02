package server

// 사고 기록 · 초안 통합 테스트: 진짜 Postgres · Valkey(testutil) + HTTPS 서버. PLAN M5 완료 기준과 M5 사전 검증 시나리오를 옮겼다.
// 예시 글은 지어낸 것이다. Docker가 없으면 로컬에서는 건너뛰고, CI에서는 실패한다.

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/sagming40/cbt-journal/internal/record"
)

type apiDraft struct {
	ID            string    `json:"id"`
	Rev           int       `json:"rev"`
	Step          int       `json:"step"`
	OccurredAt    time.Time `json:"occurred_at"`
	Situation     string    `json:"situation"`
	Emotions      []apiDEm  `json:"emotions"`
	DistortionIDs []int     `json:"distortion_ids"`
}

type apiDEm struct {
	EmotionID int  `json:"emotion_id"`
	Before    *int `json:"intensity_before"`
	After     *int `json:"intensity_after"`
}

type apiDraftList struct {
	Items []struct {
		ID               string    `json:"id"`
		Step             int       `json:"step"`
		SituationPreview string    `json:"situation_preview"`
		UpdatedAt        time.Time `json:"updated_at"`
	} `json:"items"`
}

type apiRecord struct {
	ID                 string    `json:"id"`
	OccurredAt         time.Time `json:"occurred_at"`
	Situation          string    `json:"situation"`
	AutomaticThought   string    `json:"automatic_thought"`
	EvidenceFor        string    `json:"evidence_for"`
	EvidenceAgainst    string    `json:"evidence_against"`
	AlternativeThought string    `json:"alternative_thought"`
	Emotions           []struct {
		EmotionID int    `json:"emotion_id"`
		Name      string `json:"name"`
		Before    int    `json:"intensity_before"`
		After     int    `json:"intensity_after"`
	} `json:"emotions"`
	Distortions []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"distortions"`
}

type apiRecordDay struct {
	Date  string      `json:"date"`
	Items []apiRecord `json:"items"`
}

func rem(id, before, after int) map[string]int {
	return map[string]int{"emotion_id": id, "intensity_before": before, "intensity_after": after}
}

// draftBody는 초안 저장 본문이다. 지금 쓰는 중인 모습(1단계, 상황만)으로 시작한다.
func draftBody(situation string) map[string]any {
	return map[string]any{
		"step": 1, "occurred_at": time.Now().UTC().Format(time.RFC3339),
		"situation": situation, "automatic_thought": "", "evidence_for": "", "evidence_against": "", "alternative_thought": "",
		"emotions": []any{}, "distortion_ids": []any{},
	}
}

func withRev(b map[string]any, rev int) map[string]any {
	b["rev"] = rev
	return b
}

// recordBody는 완료 본문이다 (필수 칸을 다 채운 모습).
func recordBody(ids map[string]int) map[string]any {
	return map[string]any{
		"situation": "회의에서 발표하다 말이 꼬였다", "automatic_thought": "또 내가 망쳤다",
		"evidence_for": "말이 두 번 꼬였다", "evidence_against": "질문은 잘 받았다", "alternative_thought": "실수는 했지만 끝까지 해냈다",
		"emotions": []any{rem(ids["불안"], 80, 35)}, "distortion_ids": []any{},
	}
}

func (a *testApp) newDraft(c *http.Client, situation string) apiDraft {
	a.t.Helper()
	r := a.call(c, "POST", "/api/drafts", draftBody(situation))
	if r.status != http.StatusCreated {
		a.t.Fatalf("초안 만들기: %d %s", r.status, r.body)
	}
	return decode[apiDraft](a.t, r)
}

func (a *testApp) draftCount(c *http.Client) int {
	a.t.Helper()
	r := a.call(c, "GET", "/api/drafts", nil)
	if r.status != http.StatusOK {
		a.t.Fatalf("초안 목록: %d %s", r.status, r.body)
	}
	return len(decode[apiDraftList](a.t, r).Items)
}

func TestDistortionsSeed(t *testing.T) {
	a := newApp(t)
	c := a.signedIn(testEmail)
	r := a.call(c, "GET", "/api/distortions", nil)
	list := decode[struct {
		Items []struct {
			ID          int     `json:"id"`
			Name        string  `json:"name"`
			Description *string `json:"description"`
		} `json:"items"`
	}](t, r)
	if len(list.Items) != 11 {
		t.Fatalf("왜곡 %d개, 기대 11개", len(list.Items))
	}
	if list.Items[0].Name != "흑백논리" || list.Items[10].Name != "파국화" {
		t.Errorf("순서: %q … %q", list.Items[0].Name, list.Items[10].Name)
	}
	// 문구는 사용자가 쓰기 전까지 비어 있다 (DATA_MODEL 3.3절)
	if list.Items[0].Description != nil {
		t.Errorf("설명은 null이어야 한다: %q", *list.Items[0].Description)
	}
}

// PLAN M5 완료 기준: 오래된 rev로 저장하면 409. "지금 내용으로 덮어쓰기"는 최신 rev를 읽어 저장한다.
func TestDraftRevConflict(t *testing.T) {
	a := newApp(t)
	phone := a.signedIn(testEmail)
	pc := a.device()
	if r := a.login(pc, testEmail, testPassword, false); r.status != http.StatusOK {
		t.Fatalf("로그인: %d", r.status)
	}

	d := a.newDraft(phone, "회의")
	if d.Rev != 1 || d.Step != 1 || d.Situation != "회의" {
		t.Fatalf("처음 초안: %+v", d)
	}
	if d.Emotions == nil || d.DistortionIDs == nil {
		t.Errorf("빈 배열은 null이 아니라 []여야 한다: %+v", d)
	}

	path := "/api/drafts/" + d.ID
	// PC가 먼저 저장 → rev 2
	r := a.call(pc, "PUT", path, withRev(draftBody("회의 끝나고"), 1))
	if r.status != http.StatusOK {
		t.Fatalf("저장: %d %s", r.status, r.body)
	}
	if got := decode[apiDraft](t, r); got.Rev != 2 || got.Situation != "회의 끝나고" {
		t.Errorf("저장 결과: %+v", got)
	}
	// 휴대폰은 아직 rev 1 → 409, 내용은 PC 것 그대로
	r = a.call(phone, "PUT", path, withRev(draftBody("휴대폰 글"), 1))
	if r.status != http.StatusConflict || r.errorCode() != "DRAFT_CONFLICT" {
		t.Fatalf("오래된 rev: %d %s", r.status, r.body)
	}
	cur := decode[apiDraft](t, a.call(phone, "GET", path, nil))
	if cur.Situation != "회의 끝나고" || cur.Rev != 2 {
		t.Errorf("충돌 뒤 초안이 바뀌면 안 된다: %+v", cur)
	}
	// 덮어쓰기(D-02): 최신 rev로 다시 저장
	r = a.call(phone, "PUT", path, withRev(draftBody("휴대폰 글"), cur.Rev))
	if got := decode[apiDraft](t, r); r.status != http.StatusOK || got.Rev != 3 || got.Situation != "휴대폰 글" {
		t.Errorf("덮어쓰기: %d %s", r.status, r.body)
	}
}

// 두 기기가 같은 rev로 동시에 저장하면 하나만 성공한다 (Lua로 "비교 + 저장"이 한 덩어리).
func TestDraftConcurrentSaveOneWins(t *testing.T) {
	a := newApp(t)
	c := a.signedIn(testEmail)
	d := a.newDraft(c, "처음")
	var wg sync.WaitGroup
	statuses := make([]int, 8)
	for i := range statuses {
		wg.Go(func() {
			statuses[i] = a.call(c, "PUT", "/api/drafts/"+d.ID, withRev(draftBody(fmt.Sprintf("글 %d", i)), 1)).status
		})
	}
	wg.Wait()
	ok := 0
	for _, s := range statuses {
		switch s {
		case http.StatusOK:
			ok++
		case http.StatusConflict:
		default:
			t.Errorf("상태: %d", s)
		}
	}
	if ok != 1 {
		t.Errorf("성공 %d개, 기대 1개 (%v)", ok, statuses)
	}
}

// PLAN M5 완료 기준: 11개째 거절, 만료된 초안은 개수에서 빠짐.
func TestDraftLimitAndGhosts(t *testing.T) {
	a := newApp(t)
	c := a.signedIn(testEmail)
	var first apiDraft
	for i := range 10 {
		d := a.newDraft(c, fmt.Sprintf("초안 %d", i))
		if i == 0 {
			first = d
		}
	}
	r := a.call(c, "POST", "/api/drafts", draftBody("열한 번째"))
	if r.status != http.StatusConflict || r.errorCode() != "DRAFT_LIMIT" {
		t.Fatalf("11번째: %d %s", r.status, r.body)
	}
	if n := a.draftCount(c); n != 10 {
		t.Fatalf("거절 뒤 %d개, 기대 10개", n)
	}

	// 14일이 지나 키만 사라진 초안을 흉내 낸다 — 목록(drafts:)에는 id가 남아 있다.
	uid := a.userIDOf(testEmail)
	ctx := context.Background()
	if err := a.rdb.Del(ctx, "draft:"+uid+":"+first.ID).Err(); err != nil {
		t.Fatal(err)
	}
	if n := a.rdb.ZCard(ctx, "drafts:"+uid).Val(); n != 10 {
		t.Fatalf("목록에 유령이 남아 있어야 한다: %d", n)
	}
	// 유령은 개수를 차지하지 않는다: 다시 만들 수 있다
	a.newDraft(c, "자리가 났다")
	if n := a.draftCount(c); n != 10 {
		t.Errorf("유령 정리 뒤 %d개, 기대 10개", n)
	}
	// 목록 조회만으로도 유령이 정리된다
	if err := a.rdb.Del(ctx, "draft:"+uid+":"+decode[apiDraftList](t, a.call(c, "GET", "/api/drafts", nil)).Items[0].ID).Err(); err != nil {
		t.Fatal(err)
	}
	if n := a.draftCount(c); n != 9 {
		t.Errorf("목록 조회: %d개, 기대 9개", n)
	}
	if n := a.rdb.ZCard(ctx, "drafts:"+uid).Val(); n != 9 {
		t.Errorf("목록에서 유령을 지워야 한다: %d", n)
	}
	// 없어진 초안을 저장하면 404
	if r := a.call(c, "PUT", "/api/drafts/"+first.ID, withRev(draftBody("x"), 1)); r.status != http.StatusNotFound {
		t.Errorf("없어진 초안 저장: %d", r.status)
	}
}

// 10개 제한은 동시에 만들어도 지켜진다 (Lua로 "세기 + 저장"이 한 덩어리).
func TestDraftLimitConcurrent(t *testing.T) {
	a := newApp(t)
	c := a.signedIn(testEmail)
	var wg sync.WaitGroup
	for i := range 15 {
		wg.Go(func() { a.call(c, "POST", "/api/drafts", draftBody(fmt.Sprintf("글 %d", i))) })
	}
	wg.Wait()
	if n := a.draftCount(c); n != 10 {
		t.Errorf("초안 %d개, 기대 10개", n)
	}
}

// DATA_MODEL 4.2절: 초안 키 · 목록 키 모두 TTL이 있고, 저장하면 다시 14일이 된다. TTL 없는 키를 만들지 않는다(CLAUDE.md).
func TestDraftKeysHaveTTL(t *testing.T) {
	a := newApp(t)
	c := a.signedIn(testEmail)
	d := a.newDraft(c, "처음")
	uid := a.userIDOf(testEmail)
	ctx := context.Background()
	const day = 24 * time.Hour
	for _, key := range []string{"draft:" + uid + ":" + d.ID, "drafts:" + uid} {
		a.rdb.Expire(ctx, key, day)
	}
	if r := a.call(c, "PUT", "/api/drafts/"+d.ID, withRev(draftBody("다시"), 1)); r.status != http.StatusOK {
		t.Fatalf("저장: %d %s", r.status, r.body)
	}
	for _, key := range []string{"draft:" + uid + ":" + d.ID, "drafts:" + uid} {
		if ttl := a.rdb.TTL(ctx, key).Val(); ttl < 13*day || ttl > 14*day {
			t.Errorf("%s TTL: %v", key, ttl)
		}
	}
}

func TestDraftValidation(t *testing.T) {
	a := newApp(t)
	c := a.signedIn(testEmail)
	bad := func(mut func(map[string]any), field string) {
		t.Helper()
		b := draftBody("상황")
		mut(b)
		r := a.call(c, "POST", "/api/drafts", b)
		if r.status != http.StatusBadRequest || r.fields()[field] == "" {
			t.Errorf("%s: %d %s", field, r.status, r.body)
		}
	}
	bad(func(b map[string]any) { b["step"] = 0 }, "step")
	bad(func(b map[string]any) { b["step"] = 9 }, "step")
	bad(func(b map[string]any) { delete(b, "occurred_at") }, "occurred_at")
	bad(func(b map[string]any) { b["situation"] = strings.Repeat("가", 1001) }, "situation")
	bad(func(b map[string]any) { b["evidence_for"] = strings.Repeat("가", 2001) }, "evidence_for")
	bad(func(b map[string]any) { b["emotions"] = []any{rem(1, 101, 0)} }, "emotions")
	bad(func(b map[string]any) { b["emotions"] = []any{rem(1, 10, 10), rem(1, 20, 20)} }, "emotions")
	bad(func(b map[string]any) { b["distortion_ids"] = []any{1, 1} }, "distortion_ids")

	// 빈 칸 · 후 강도 없음은 초안에서 괜찮고, 쓰는 중인 공백은 다듬지 않는다
	b := draftBody("상황 ")
	b["emotions"] = []any{map[string]any{"emotion_id": 1, "intensity_before": 60, "intensity_after": nil}}
	r := a.call(c, "POST", "/api/drafts", b)
	if r.status != http.StatusCreated {
		t.Fatalf("느슨한 초안: %d %s", r.status, r.body)
	}
	d := decode[apiDraft](t, r)
	if d.Situation != "상황 " || d.Emotions[0].After != nil {
		t.Errorf("초안: %+v", d)
	}
}

func TestDraftListAndDelete(t *testing.T) {
	a := newApp(t)
	c := a.signedIn(testEmail)
	base := time.Now().UTC().Truncate(time.Second)
	a.now = base
	older := a.newDraft(c, strings.Repeat("긴", 50))
	a.now = base.Add(time.Minute)
	a.newDraft(c, "나중 초안")

	list := decode[apiDraftList](t, a.call(c, "GET", "/api/drafts", nil))
	if len(list.Items) != 2 || list.Items[0].SituationPreview != "나중 초안" {
		t.Fatalf("최근 저장순: %+v", list)
	}
	if got := []rune(list.Items[1].SituationPreview); len(got) != 40 {
		t.Errorf("앞부분은 40자: %d", len(got))
	}
	// 옛 초안을 저장하면 맨 위로 온다
	a.now = base.Add(2 * time.Minute)
	if r := a.call(c, "PUT", "/api/drafts/"+older.ID, withRev(draftBody("다시 쓴 글"), 1)); r.status != http.StatusOK {
		t.Fatalf("저장: %d", r.status)
	}
	list = decode[apiDraftList](t, a.call(c, "GET", "/api/drafts", nil))
	if list.Items[0].ID != older.ID {
		t.Errorf("저장한 초안이 위로: %+v", list)
	}

	if r := a.call(c, "DELETE", "/api/drafts/"+older.ID, nil); r.status != http.StatusNoContent {
		t.Fatalf("지우기: %d", r.status)
	}
	if r := a.call(c, "GET", "/api/drafts/"+older.ID, nil); r.status != http.StatusNotFound {
		t.Errorf("지운 초안: %d", r.status)
	}
	if r := a.call(c, "DELETE", "/api/drafts/"+older.ID, nil); r.status != http.StatusNotFound {
		t.Errorf("두 번 지우기: %d", r.status)
	}
	if r := a.call(c, "GET", "/api/drafts/아무말", nil); r.status != http.StatusNotFound {
		t.Errorf("이상한 id: %d", r.status)
	}
}

// PLAN M5 완료 기준: 남의 기록 · 초안 404.
func TestRecordAndDraftOtherUser404(t *testing.T) {
	a := newApp(t)
	alice := a.signedIn(testEmail)
	bob := a.signedIn("bob@example.com")
	ids := a.emotionIDs(alice)

	d := a.newDraft(alice, "앨리스 초안")
	draftPath := "/api/drafts/" + d.ID
	for _, tc := range []struct {
		method, path string
		body         any
	}{
		{"GET", draftPath, nil},
		{"PUT", draftPath, withRev(draftBody("가로채기"), 1)},
		{"DELETE", draftPath, nil},
	} {
		if r := a.call(bob, tc.method, tc.path, tc.body); r.status != http.StatusNotFound || r.errorCode() != "NOT_FOUND" {
			t.Errorf("%s 남의 초안: %d %s", tc.method, r.status, r.body)
		}
	}
	if n := a.draftCount(bob); n != 0 {
		t.Errorf("밥의 초안 목록에 남의 것이 보인다: %d", n)
	}
	if got := decode[apiDraft](t, a.call(alice, "GET", draftPath, nil)); got.Situation != "앨리스 초안" || got.Rev != 1 {
		t.Errorf("남이 건드리면 안 된다: %+v", got)
	}

	// 밥이 앨리스의 초안 id로 완료해도 앨리스의 초안은 지워지지 않는다 (키에 user_id)
	body := recordBody(ids)
	body["draft_id"] = d.ID
	if r := a.call(bob, "POST", "/api/records", body); r.status != http.StatusCreated {
		t.Fatalf("밥 완료: %d %s", r.status, r.body)
	}
	if r := a.call(alice, "GET", draftPath, nil); r.status != http.StatusOK {
		t.Errorf("앨리스 초안이 남아 있어야 한다: %d", r.status)
	}

	// 앨리스도 같은 초안 id로 완료할 수 있다 — UNIQUE가 사용자별이라 밥의 기록과 부딪히지 않는다
	r := a.call(alice, "POST", "/api/records", body)
	if r.status != http.StatusCreated {
		t.Fatalf("앨리스 완료: %d %s", r.status, r.body)
	}
	rec := decode[apiRecord](t, r)
	path := "/api/records/" + rec.ID
	for _, tc := range []struct {
		method string
		body   any
	}{
		{"GET", nil},
		{"PUT", recordBody(ids)},
		{"DELETE", nil},
	} {
		if r := a.call(bob, tc.method, path, tc.body); r.status != http.StatusNotFound || r.errorCode() != "NOT_FOUND" {
			t.Errorf("%s 남의 기록: %d %s", tc.method, r.status, r.body)
		}
	}
	if got := decode[apiRecord](t, a.call(alice, "GET", path, nil)); got.Situation != rec.Situation {
		t.Errorf("남이 건드리면 안 된다: %+v", got)
	}
}

func TestRecordCompleteRemovesDraft(t *testing.T) {
	a := newApp(t)
	c := a.signedIn(testEmail)
	ids := a.emotionIDs(c)
	dist := a.distortionIDs(c)
	d := a.newDraft(c, "상황")

	body := recordBody(ids)
	body["draft_id"] = d.ID
	// 고른 순서와 상관없이 화면 순서로 돌아온다
	body["emotions"] = []any{rem(ids["짜증"], 40, 40), rem(ids["불안"], 80, 35)}
	body["distortion_ids"] = []any{dist["파국화"], dist["흑백논리"]}
	body["situation"] = "  회의에서 말이 꼬였다  "
	occurred := time.Now().Add(-48 * time.Hour).UTC().Truncate(time.Second)
	body["occurred_at"] = occurred.Format(time.RFC3339)

	verBefore := a.statsVer()
	r := a.call(c, "POST", "/api/records", body)
	if r.status != http.StatusCreated {
		t.Fatalf("완료: %d %s", r.status, r.body)
	}
	rec := decode[apiRecord](t, r)
	if rec.Situation != "회의에서 말이 꼬였다" {
		t.Errorf("완료할 때는 앞뒤 공백을 다듬는다: %q", rec.Situation)
	}
	if !rec.OccurredAt.Equal(occurred) {
		t.Errorf("일어난 시각(T6): %v, 기대 %v", rec.OccurredAt, occurred)
	}
	if len(rec.Emotions) != 2 || rec.Emotions[0].Name != "불안" || rec.Emotions[0].Before != 80 || rec.Emotions[0].After != 35 || rec.Emotions[1].Name != "짜증" {
		t.Errorf("감정: %+v", rec.Emotions)
	}
	if len(rec.Distortions) != 2 || rec.Distortions[0].Name != "흑백논리" || rec.Distortions[1].Name != "파국화" {
		t.Errorf("왜곡: %+v", rec.Distortions)
	}
	// 커밋 뒤에 초안이 사라진다 · 통계 버전이 오른다 (DATA_MODEL 4.3 · 4.4절)
	if r := a.call(c, "GET", "/api/drafts/"+d.ID, nil); r.status != http.StatusNotFound {
		t.Errorf("완료한 초안: %d", r.status)
	}
	if n := a.draftCount(c); n != 0 {
		t.Errorf("초안 목록: %d", n)
	}
	if v := a.statsVer(); v != verBefore+1 {
		t.Errorf("통계 버전: %d → %d", verBefore, v)
	}

	got := decode[apiRecord](t, a.call(c, "GET", "/api/records/"+rec.ID, nil))
	if got.EvidenceFor != "말이 두 번 꼬였다" || got.AlternativeThought == "" || len(got.Emotions) != 2 {
		t.Errorf("다시 읽기: %+v", got)
	}
}

// 같은 초안을 두 번 완료해도 기록은 하나 (DATA_MODEL 3.5절): 버튼 연타 · 네트워크 재시도.
func TestRecordCompleteIdempotent(t *testing.T) {
	a := newApp(t)
	c := a.signedIn(testEmail)
	ids := a.emotionIDs(c)
	d := a.newDraft(c, "상황")
	body := recordBody(ids)
	body["draft_id"] = d.ID

	first := a.call(c, "POST", "/api/records", body)
	if first.status != http.StatusCreated {
		t.Fatalf("처음: %d %s", first.status, first.body)
	}
	verAfterFirst := a.statsVer()
	second := a.call(c, "POST", "/api/records", body)
	if second.status != http.StatusOK {
		t.Fatalf("다시: %d %s", second.status, second.body)
	}
	if decode[apiRecord](t, first).ID != decode[apiRecord](t, second).ID {
		t.Error("같은 기록을 돌려줘야 한다")
	}
	if got := decode[apiRecord](t, second); len(got.Emotions) != 1 {
		t.Errorf("감정이 그대로여야 한다: %+v", got.Emotions)
	}
	if v := a.statsVer(); v != verAfterFirst {
		t.Errorf("새로 만든 게 없으니 통계 버전 그대로: %d → %d", verAfterFirst, v)
	}
	if n := a.countRecords(); n != 1 {
		t.Errorf("기록 %d개, 기대 1개", n)
	}

	// 동시에 와도 하나
	d2 := a.newDraft(c, "연타")
	body["draft_id"] = d2.ID
	var wg sync.WaitGroup
	for range 6 {
		wg.Go(func() {
			if r := a.call(c, "POST", "/api/records", body); r.status != http.StatusCreated && r.status != http.StatusOK {
				t.Errorf("동시 완료: %d %s", r.status, r.body)
			}
		})
	}
	wg.Wait()
	if n := a.countRecords(); n != 2 {
		t.Errorf("동시 완료 뒤 기록 %d개, 기대 2개", n)
	}
	// 초안 없이 완료는 매번 새 기록 (NULL은 UNIQUE에 걸리지 않는다)
	delete(body, "draft_id")
	for range 2 {
		if r := a.call(c, "POST", "/api/records", body); r.status != http.StatusCreated {
			t.Errorf("초안 없이: %d %s", r.status, r.body)
		}
	}
	if n := a.countRecords(); n != 4 {
		t.Errorf("기록 %d개, 기대 4개", n)
	}
}

// PLAN M5 완료 기준: 완료가 실패하면 초안이 남는다 — 입력 오류와, 트랜잭션 도중 DB 오류 둘 다.
func TestRecordCompleteFailureKeepsDraft(t *testing.T) {
	a := newApp(t)
	c := a.signedIn(testEmail)
	ids := a.emotionIDs(c)
	dist := a.distortionIDs(c)
	d := a.newDraft(c, "상황")

	body := recordBody(ids)
	body["draft_id"] = d.ID
	body["alternative_thought"] = "  "
	r := a.call(c, "POST", "/api/records", body)
	if r.status != http.StatusBadRequest || r.fields()["alternative_thought"] == "" {
		t.Fatalf("입력 오류: %d %s", r.status, r.body)
	}
	if r := a.call(c, "GET", "/api/drafts/"+d.ID, nil); r.status != http.StatusOK {
		t.Errorf("입력 오류 뒤 초안: %d", r.status)
	}

	// 본문 INSERT 뒤, 왜곡 행을 넣다가 DB가 실패하는 경우를 트리거로 만든다.
	ctx := context.Background()
	if _, err := a.pool.Exec(ctx, `
		CREATE FUNCTION test_fail() RETURNS trigger AS $$ BEGIN RAISE EXCEPTION 'test'; END $$ LANGUAGE plpgsql;
		CREATE TRIGGER test_fail BEFORE INSERT ON record_distortions FOR EACH ROW EXECUTE FUNCTION test_fail();`); err != nil {
		t.Fatal(err)
	}
	drop := func() {
		_, _ = a.pool.Exec(ctx, `DROP TRIGGER IF EXISTS test_fail ON record_distortions; DROP FUNCTION IF EXISTS test_fail();`)
	}
	t.Cleanup(drop)

	body = recordBody(ids)
	body["draft_id"] = d.ID
	body["distortion_ids"] = []any{dist["파국화"]}
	if r := a.call(c, "POST", "/api/records", body); r.status != http.StatusInternalServerError {
		t.Fatalf("DB 실패: %d %s", r.status, r.body)
	}
	if r := a.call(c, "GET", "/api/drafts/"+d.ID, nil); r.status != http.StatusOK {
		t.Errorf("DB 실패 뒤 초안: %d", r.status)
	}
	if n := a.countRecords(); n != 0 {
		t.Errorf("되돌려져야 한다: 기록 %d개", n)
	}

	// 고치고 다시 완료하면 된다
	drop()
	if r := a.call(c, "POST", "/api/records", body); r.status != http.StatusCreated {
		t.Fatalf("다시 완료: %d %s", r.status, r.body)
	}
}

func TestRecordValidation(t *testing.T) {
	a := newApp(t)
	c := a.signedIn(testEmail)
	ids := a.emotionIDs(c)
	dist := a.distortionIDs(c)
	a.now = time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)

	bad := func(mut func(map[string]any), field string) {
		t.Helper()
		b := recordBody(ids)
		mut(b)
		r := a.call(c, "POST", "/api/records", b)
		if r.status != http.StatusBadRequest || r.errorCode() != "INVALID_INPUT" || r.fields()[field] == "" {
			t.Errorf("%s: %d %s", field, r.status, r.body)
		}
	}
	bad(func(b map[string]any) { b["situation"] = "" }, "situation")
	bad(func(b map[string]any) { b["automatic_thought"] = "   " }, "automatic_thought")
	bad(func(b map[string]any) { b["alternative_thought"] = "" }, "alternative_thought")
	bad(func(b map[string]any) { b["situation"] = strings.Repeat("가", 1001) }, "situation")
	bad(func(b map[string]any) { b["evidence_against"] = strings.Repeat("가", 2001) }, "evidence_against")
	bad(func(b map[string]any) { b["emotions"] = []any{} }, "emotions")
	bad(func(b map[string]any) { b["emotions"] = []any{rem(9999, 10, 10)} }, "emotions")
	bad(func(b map[string]any) { b["emotions"] = []any{rem(ids["불안"], 10, 101)} }, "emotions")
	bad(func(b map[string]any) { b["emotions"] = []any{rem(ids["불안"], 10, 10), rem(ids["불안"], 20, 20)} }, "emotions")
	bad(func(b map[string]any) {
		b["emotions"] = []any{map[string]any{"emotion_id": ids["불안"], "intensity_before": 10, "intensity_after": nil}}
	}, "emotions")
	bad(func(b map[string]any) { b["distortion_ids"] = []any{9999} }, "distortion_ids")
	bad(func(b map[string]any) { b["distortion_ids"] = []any{dist["파국화"], dist["파국화"]} }, "distortion_ids")
	bad(func(b map[string]any) { b["occurred_at"] = a.now.Add(6 * time.Minute).Format(time.RFC3339) }, "occurred_at")
	bad(func(b map[string]any) { b["draft_id"] = "아무말" }, "draft_id")

	// 경계: 딱 1000자 · 2000자 · 앞으로 5분 · 왜곡 0개 · 강도 0은 된다 (DB CHECK도 통과)
	ok := recordBody(ids)
	ok["situation"] = strings.Repeat("가", 1000)
	ok["evidence_for"] = strings.Repeat("가", 2000)
	ok["evidence_against"] = ""
	ok["occurred_at"] = a.now.Add(5 * time.Minute).Format(time.RFC3339)
	ok["emotions"] = []any{rem(ids["불안"], 0, 0)}
	if r := a.call(c, "POST", "/api/records", ok); r.status != http.StatusCreated {
		t.Errorf("경계값: %d %s", r.status, r.body)
	}
}

// 수정은 통째로, 나중에 저장한 쪽이 이긴다 (DATA_MODEL 3.5절). 지워진 걸 고치면 404.
func TestRecordUpdateDelete(t *testing.T) {
	a := newApp(t)
	c := a.signedIn(testEmail)
	ids := a.emotionIDs(c)
	dist := a.distortionIDs(c)
	create := recordBody(ids)
	create["distortion_ids"] = []any{dist["파국화"]}
	rec := decode[apiRecord](t, a.call(c, "POST", "/api/records", create))
	path := "/api/records/" + rec.ID

	edit := recordBody(ids)
	edit["situation"] = "고친 상황"
	edit["emotions"] = []any{rem(ids["슬픔"], 50, 20)}
	edit["distortion_ids"] = []any{dist["개인화"]}
	ver := a.statsVer()
	r := a.call(c, "PUT", path, edit)
	if r.status != http.StatusOK {
		t.Fatalf("수정: %d %s", r.status, r.body)
	}
	got := decode[apiRecord](t, r)
	if got.Situation != "고친 상황" || len(got.Emotions) != 1 || got.Emotions[0].Name != "슬픔" || len(got.Distortions) != 1 || got.Distortions[0].Name != "개인화" {
		t.Errorf("수정 결과: %+v", got)
	}
	if v := a.statsVer(); v != ver+1 {
		t.Errorf("수정하면 통계 버전: %d → %d", ver, v)
	}

	edit["situation"] = "다른 기기에서 나중에 저장"
	if r := a.call(c, "PUT", path, edit); r.status != http.StatusOK {
		t.Fatalf("나중 저장: %d", r.status)
	}
	if got := decode[apiRecord](t, a.call(c, "GET", path, nil)); got.Situation != "다른 기기에서 나중에 저장" {
		t.Errorf("나중에 저장한 쪽이 이긴다: %q", got.Situation)
	}

	// 잘못된 수정은 아무것도 바꾸지 않는다
	broken := recordBody(ids)
	broken["emotions"] = []any{}
	if r := a.call(c, "PUT", path, broken); r.status != http.StatusBadRequest {
		t.Errorf("잘못된 수정: %d", r.status)
	}
	if got := decode[apiRecord](t, a.call(c, "GET", path, nil)); len(got.Emotions) != 1 {
		t.Errorf("감정이 남아 있어야 한다: %+v", got.Emotions)
	}

	ver = a.statsVer()
	if r := a.call(c, "DELETE", path, nil); r.status != http.StatusNoContent {
		t.Fatalf("지우기: %d", r.status)
	}
	if v := a.statsVer(); v != ver+1 {
		t.Errorf("지우면 통계 버전: %d → %d", ver, v)
	}
	if r := a.call(c, "GET", path, nil); r.status != http.StatusNotFound {
		t.Errorf("지운 기록: %d", r.status)
	}
	if r := a.call(c, "PUT", path, edit); r.status != http.StatusNotFound {
		t.Errorf("지운 기록 수정: %d", r.status)
	}
	if r := a.call(c, "DELETE", path, nil); r.status != http.StatusNotFound {
		t.Errorf("두 번 지우기: %d", r.status)
	}
	// 자식 행도 같이 사라진다 (CASCADE)
	var n int
	if err := a.pool.QueryRow(context.Background(), `SELECT (SELECT count(*) FROM record_emotions) + (SELECT count(*) FROM record_distortions)`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("자식 행 %d개 남음", n)
	}
}

// "그날"은 사용자 시간대 기준 (DATA_MODEL 3.4절). 서울 11월 2일 0시 = UTC 11월 1일 15시.
func TestRecordDayFollowsTimezone(t *testing.T) {
	a := newApp(t)
	c := a.signedIn(testEmail)
	ids := a.emotionIDs(c)
	a.now = time.Date(2026, 11, 2, 3, 0, 0, 0, time.UTC) // 서울 12시
	put := func(at time.Time, s string) {
		t.Helper()
		b := recordBody(ids)
		b["situation"] = s
		b["occurred_at"] = at.Format(time.RFC3339)
		if r := a.call(c, "POST", "/api/records", b); r.status != http.StatusCreated {
			t.Fatalf("%s: %d %s", s, r.status, r.body)
		}
	}
	put(time.Date(2026, 11, 1, 14, 59, 0, 0, time.UTC), "어제 끝")
	put(time.Date(2026, 11, 1, 15, 0, 0, 0, time.UTC), "오늘 시작")
	put(time.Date(2026, 11, 2, 2, 0, 0, 0, time.UTC), "오늘 낮")

	day := decode[apiRecordDay](t, a.call(c, "GET", "/api/records", nil))
	if day.Date != "2026-11-02" || len(day.Items) != 2 || day.Items[0].Situation != "오늘 시작" || day.Items[1].Situation != "오늘 낮" {
		t.Errorf("오늘(시간순): %+v", day)
	}
	day = decode[apiRecordDay](t, a.call(c, "GET", "/api/records?date=2026-11-01", nil))
	if len(day.Items) != 1 || day.Items[0].Situation != "어제 끝" {
		t.Errorf("어제: %+v", day)
	}
	if r := a.call(c, "GET", "/api/records?date=어제", nil); r.status != http.StatusBadRequest {
		t.Errorf("이상한 날짜: %d", r.status)
	}
	// 시간대를 바꾸면 "어제 끝"(UTC 14:59)이 UTC의 11월 1일로 남고 "오늘 시작"도 같은 날이 된다
	if r := a.call(c, "PATCH", "/api/me", map[string]string{"timezone": "UTC"}); r.status != http.StatusOK {
		t.Fatalf("시간대: %d", r.status)
	}
	day = decode[apiRecordDay](t, a.call(c, "GET", "/api/records?date=2026-11-01", nil))
	if len(day.Items) != 2 {
		t.Errorf("UTC 기준 11월 1일: %+v", day)
	}
}

// Valkey가 죽어 있으면 초안만 503이고, 완료 · 조회는 된다 (DATA_MODEL 4.3절, decisions 0005).
func TestDraftsUnavailableWhenValkeyDown(t *testing.T) {
	a := newApp(t)
	c := a.signedIn(testEmail)
	ids := a.emotionIDs(c)

	dead := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: -1, DialTimeout: 100 * time.Millisecond})
	t.Cleanup(func() { _ = dead.Close() })
	discard := slog.New(slog.DiscardHandler)
	svc := record.NewService(record.Options{Pool: a.pool, Valkey: dead, Logger: discard})
	srv := httptest.NewTLSServer(New(Deps{Logger: discard, Checks: Checks{Postgres: okCheck, Valkey: okCheck}, Web: fstest.MapFS{}, Auth: a.svc, Record: svc}))
	t.Cleanup(srv.Close)
	a.srv = srv

	const someID = "01900000-0000-7000-8000-000000000000"
	for _, tc := range []struct {
		method, path string
		body         any
	}{
		{"GET", "/api/drafts", nil},
		{"POST", "/api/drafts", draftBody("상황")},
		{"GET", "/api/drafts/" + someID, nil},
		{"PUT", "/api/drafts/" + someID, withRev(draftBody("상황"), 1)},
		{"DELETE", "/api/drafts/" + someID, nil},
	} {
		if r := a.call(c, tc.method, tc.path, tc.body); r.status != http.StatusServiceUnavailable || r.errorCode() != "DRAFTS_UNAVAILABLE" {
			t.Errorf("%s %s: %d %s", tc.method, tc.path, r.status, r.body)
		}
	}
	body := recordBody(ids)
	body["draft_id"] = someID
	r := a.call(c, "POST", "/api/records", body)
	if r.status != http.StatusCreated {
		t.Fatalf("Valkey 없이 완료: %d %s", r.status, r.body)
	}
	if r := a.call(c, "GET", "/api/records/"+decode[apiRecord](t, r).ID, nil); r.status != http.StatusOK {
		t.Errorf("조회: %d", r.status)
	}
}

func (a *testApp) distortionIDs(c *http.Client) map[string]int {
	a.t.Helper()
	list := decode[struct {
		Items []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}](a.t, a.call(c, "GET", "/api/distortions", nil))
	m := map[string]int{}
	for _, d := range list.Items {
		m[d.Name] = d.ID
	}
	return m
}

func (a *testApp) userIDOf(email string) string {
	a.t.Helper()
	var id string
	if err := a.pool.QueryRow(context.Background(), `SELECT id::text FROM users WHERE email = $1`, email).Scan(&id); err != nil {
		a.t.Fatal(err)
	}
	return id
}

// statsVer는 testEmail 사용자의 통계 캐시 버전이다.
func (a *testApp) statsVer() int {
	a.t.Helper()
	v, err := a.rdb.Get(context.Background(), "stats:ver:"+a.userIDOf(testEmail)).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		a.t.Fatal(err)
	}
	return v
}

func (a *testApp) countRecords() int {
	a.t.Helper()
	var n int
	if err := a.pool.QueryRow(context.Background(), `SELECT count(*) FROM thought_records`).Scan(&n); err != nil {
		a.t.Fatal(err)
	}
	return n
}
