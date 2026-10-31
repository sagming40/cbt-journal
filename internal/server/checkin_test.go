package server

// 체크인 통합 테스트: 진짜 Postgres · Valkey(testutil) + HTTPS 서버. PLAN M4 완료 기준과 M4 사전 검증 시나리오를 옮겼다.
// Docker가 없으면 로컬에서는 건너뛰고, CI에서는 실패한다.

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

type apiEmotion struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type apiCheckin struct {
	ID         string    `json:"id"`
	OccurredAt time.Time `json:"occurred_at"`
	Note       string    `json:"note"`
	Emotions   []struct {
		EmotionID int    `json:"emotion_id"`
		Name      string `json:"name"`
		Intensity int    `json:"intensity"`
	} `json:"emotions"`
}

type apiDay struct {
	Date  string       `json:"date"`
	Items []apiCheckin `json:"items"`
}

func decode[T any](t *testing.T, r response) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(r.body, &v); err != nil {
		t.Fatalf("응답 해석: %v (%d %s)", err, r.status, r.body)
	}
	return v
}

// signedIn은 사용자를 만들고 로그인한 기기 하나를 돌려준다.
func (a *testApp) signedIn(email string) *http.Client {
	a.t.Helper()
	a.createUser(email, testPassword)
	c := a.device()
	if r := a.login(c, email, testPassword, false); r.status != http.StatusOK {
		a.t.Fatalf("로그인: %d %s", r.status, r.body)
	}
	return c
}

// emotionIDs는 감정 이름 → id다. 테스트가 seed의 id 번호에 기대지 않게 API로 받아 온다.
func (a *testApp) emotionIDs(c *http.Client) map[string]int {
	a.t.Helper()
	r := a.call(c, "GET", "/api/emotions", nil)
	if r.status != http.StatusOK {
		a.t.Fatalf("감정 목록: %d %s", r.status, r.body)
	}
	list := decode[struct{ Items []apiEmotion }](a.t, r)
	m := map[string]int{}
	for _, e := range list.Items {
		m[e.Name] = e.ID
	}
	return m
}

func em(id, intensity int) map[string]int {
	return map[string]int{"emotion_id": id, "intensity": intensity}
}

func TestEmotionsSeed(t *testing.T) {
	a := newApp(t)
	c := a.signedIn(testEmail)
	r := a.call(c, "GET", "/api/emotions", nil)
	list := decode[struct{ Items []apiEmotion }](t, r)
	var names []string
	for _, e := range list.Items {
		names = append(names, e.Name)
	}
	// DATA_MODEL 3.3절 seed 그대로, 화면 순서대로
	want := "불안 두려움 슬픔 무기력 분노 짜증 수치심 죄책감 외로움 좌절"
	if got := strings.Join(names, " "); got != want {
		t.Errorf("감정: %q, 기대 %q", got, want)
	}
}

func TestCheckinCRUD(t *testing.T) {
	a := newApp(t)
	c := a.signedIn(testEmail)
	ids := a.emotionIDs(c)

	// 만들기: 감정 순서는 고른 순서가 아니라 화면 순서(불안 → 짜증)로 돌아온다
	r := a.call(c, "POST", "/api/checkins", map[string]any{
		"note":     "  회의 전  ",
		"emotions": []any{em(ids["짜증"], 30), em(ids["불안"], 60)},
	})
	if r.status != http.StatusCreated {
		t.Fatalf("만들기: %d %s", r.status, r.body)
	}
	created := decode[apiCheckin](t, r)
	if created.Note != "회의 전" {
		t.Errorf("메모 앞뒤 공백을 다듬어야 한다: %q", created.Note)
	}
	if len(created.Emotions) != 2 || created.Emotions[0].Name != "불안" || created.Emotions[0].Intensity != 60 || created.Emotions[1].Name != "짜증" {
		t.Errorf("감정: %+v", created.Emotions)
	}
	if time.Since(created.OccurredAt) > time.Minute || created.OccurredAt.Location() != time.UTC {
		t.Errorf("시각을 안 주면 지금(UTC): %v", created.OccurredAt)
	}

	// 오늘 목록에 나온다 (C2)
	day := decode[apiDay](t, a.call(c, "GET", "/api/checkins", nil))
	if len(day.Items) != 1 || day.Items[0].ID != created.ID {
		t.Fatalf("오늘 목록: %+v", day)
	}
	if want := time.Now().In(mustLoc(t, "Asia/Seoul")).Format("2006-01-02"); day.Date != want {
		t.Errorf("날짜를 안 주면 사용자 시간대의 오늘: %s, 기대 %s", day.Date, want)
	}

	// 하나 보기
	if r := a.call(c, "GET", "/api/checkins/"+created.ID, nil); r.status != http.StatusOK || decode[apiCheckin](t, r).ID != created.ID {
		t.Errorf("하나 보기: %d %s", r.status, r.body)
	}

	// 고치기: 통째로 바뀐다 (감정 2개 → 1개, 메모 지움)
	r = a.call(c, "PUT", "/api/checkins/"+created.ID, map[string]any{"emotions": []any{em(ids["슬픔"], 80)}})
	if r.status != http.StatusOK {
		t.Fatalf("고치기: %d %s", r.status, r.body)
	}
	updated := decode[apiCheckin](t, r)
	if updated.Note != "" || len(updated.Emotions) != 1 || updated.Emotions[0].Name != "슬픔" || updated.Emotions[0].Intensity != 80 {
		t.Errorf("고친 결과: %+v", updated)
	}

	// 지우기 → 다시 보면 404, 두 번 지우면 404
	if r := a.call(c, "DELETE", "/api/checkins/"+created.ID, nil); r.status != http.StatusNoContent {
		t.Fatalf("지우기: %d %s", r.status, r.body)
	}
	if r := a.call(c, "GET", "/api/checkins/"+created.ID, nil); r.status != http.StatusNotFound || r.errorCode() != "NOT_FOUND" {
		t.Errorf("지운 뒤 보기: %d %s", r.status, r.body)
	}
	if r := a.call(c, "DELETE", "/api/checkins/"+created.ID, nil); r.status != http.StatusNotFound {
		t.Errorf("두 번 지우기: %d", r.status)
	}
	// 감정 행도 CASCADE로 같이 지워졌다
	var left int
	if err := a.pool.QueryRow(context.Background(), "SELECT count(*) FROM checkin_emotions").Scan(&left); err != nil || left != 0 {
		t.Errorf("남은 감정 행: %d %v", left, err)
	}
}

func TestCheckinValidation(t *testing.T) {
	a := newApp(t)
	c := a.signedIn(testEmail)
	ids := a.emotionIDs(c)
	a.now = time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)

	bad := []struct {
		name  string
		body  map[string]any
		field string
	}{
		{"감정 0개", map[string]any{"emotions": []any{}}, "emotions"},
		{"감정 칸 없음", map[string]any{}, "emotions"},
		{"강도 101", map[string]any{"emotions": []any{em(ids["불안"], 101)}}, "emotions"},
		{"강도 -1", map[string]any{"emotions": []any{em(ids["불안"], -1)}}, "emotions"},
		{"같은 감정 두 번", map[string]any{"emotions": []any{em(ids["불안"], 10), em(ids["불안"], 20)}}, "emotions"},
		{"없는 감정", map[string]any{"emotions": []any{em(9999, 10)}}, "emotions"},
		{"메모 201자", map[string]any{"note": strings.Repeat("가", 201), "emotions": []any{em(ids["불안"], 10)}}, "note"},
		{"6분 뒤", map[string]any{"occurred_at": "2026-11-01T12:06:00Z", "emotions": []any{em(ids["불안"], 10)}}, "occurred_at"},
	}
	for _, b := range bad {
		r := a.call(c, "POST", "/api/checkins", b.body)
		if r.status != http.StatusBadRequest || r.errorCode() != "INVALID_INPUT" || r.fields()[b.field] == "" {
			t.Errorf("%s: %d %s, 400 + fields.%s여야 한다", b.name, r.status, r.body, b.field)
		}
	}

	// 경계값은 받는다: 강도 0 · 100, 메모 200자(한글), 4분 뒤(기기 시계 차이)
	good := map[string]any{
		"note":        strings.Repeat("가", 200),
		"occurred_at": "2026-11-01T12:04:00Z",
		"emotions":    []any{em(ids["불안"], 0), em(ids["좌절"], 100)},
	}
	if r := a.call(c, "POST", "/api/checkins", good); r.status != http.StatusCreated {
		t.Errorf("경계값: %d %s, 201이어야 한다", r.status, r.body)
	}

	// 아무것도 안 들어가야 한다 — 400이 난 만들기는 반쪽 행도 남기지 않는다
	var n int
	if err := a.pool.QueryRow(context.Background(), "SELECT count(*) FROM checkins").Scan(&n); err != nil || n != 1 {
		t.Errorf("체크인 행: %d %v, 1이어야 한다", n, err)
	}

	if r := a.call(c, "GET", "/api/checkins?date=2026-13-40", nil); r.status != http.StatusBadRequest {
		t.Errorf("틀린 날짜: %d %s, 400이어야 한다", r.status, r.body)
	}
}

// 남의 데이터는 404 (CLAUDE.md 코드 규칙, ARCHITECTURE 2절) — 없는 것과 같은 응답.
func TestCheckinOtherUser404(t *testing.T) {
	a := newApp(t)
	owner := a.signedIn("owner@example.com")
	other := a.signedIn("other@example.com")
	ids := a.emotionIDs(owner)

	r := a.call(owner, "POST", "/api/checkins", map[string]any{"emotions": []any{em(ids["불안"], 50)}})
	if r.status != http.StatusCreated {
		t.Fatalf("만들기: %d %s", r.status, r.body)
	}
	id := decode[apiCheckin](t, r).ID

	if r := a.call(other, "GET", "/api/checkins/"+id, nil); r.status != http.StatusNotFound || r.errorCode() != "NOT_FOUND" {
		t.Errorf("남의 것 보기: %d %s", r.status, r.body)
	}
	if r := a.call(other, "PUT", "/api/checkins/"+id, map[string]any{"emotions": []any{em(ids["슬픔"], 90)}}); r.status != http.StatusNotFound {
		t.Errorf("남의 것 고치기: %d %s", r.status, r.body)
	}
	if r := a.call(other, "DELETE", "/api/checkins/"+id, nil); r.status != http.StatusNotFound {
		t.Errorf("남의 것 지우기: %d %s", r.status, r.body)
	}
	if day := decode[apiDay](t, a.call(other, "GET", "/api/checkins", nil)); len(day.Items) != 0 {
		t.Errorf("남의 체크인이 내 오늘 목록에 섞였다: %+v", day.Items)
	}
	// 주인 것은 그대로다 (고치기 · 지우기가 아무것도 안 바꿈)
	mine := decode[apiCheckin](t, a.call(owner, "GET", "/api/checkins/"+id, nil))
	if len(mine.Emotions) != 1 || mine.Emotions[0].Name != "불안" || mine.Emotions[0].Intensity != 50 {
		t.Errorf("주인 체크인이 바뀌었다: %+v", mine)
	}

	// id 모양이 틀려도 404 (400이면 "그런 모양은 없다"를 알려 주게 된다)
	if r := a.call(owner, "GET", "/api/checkins/not-a-uuid", nil); r.status != http.StatusNotFound {
		t.Errorf("틀린 id 모양: %d %s", r.status, r.body)
	}
}

// "그날"은 사용자 시간대 기준 (PLAN M4 완료 기준, DATA_MODEL 3.4절).
func TestCheckinDayFollowsTimezone(t *testing.T) {
	a := newApp(t)
	c := a.signedIn(testEmail)
	ids := a.emotionIDs(c)
	a.now = time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)

	post := func(at string) string {
		t.Helper()
		r := a.call(c, "POST", "/api/checkins", map[string]any{"occurred_at": at, "emotions": []any{em(ids["불안"], 50)}})
		if r.status != http.StatusCreated {
			t.Fatalf("만들기 %s: %d %s", at, r.status, r.body)
		}
		return decode[apiCheckin](t, r).ID
	}
	day := func(date string) []string {
		t.Helper()
		d := decode[apiDay](t, a.call(c, "GET", "/api/checkins?date="+date, nil))
		var out []string
		for _, it := range d.Items {
			out = append(out, it.ID)
		}
		return out
	}

	// 15:30Z = 서울 11월 2일 00:30 · 뉴욕 11월 1일 10:30(이날 06:00Z에 서머타임이 끝나 EST) — 서울 기준 자정 근처 체크인
	nearMidnight := post("2026-11-01T15:30:00Z")
	if got := day("2026-11-02"); len(got) != 1 || got[0] != nearMidnight {
		t.Errorf("서울: 11월 2일에 있어야 한다: %v", got)
	}
	if got := day("2026-11-01"); len(got) != 0 {
		t.Errorf("서울: 11월 1일에는 없어야 한다: %v", got)
	}

	// 시간대를 뉴욕으로 바꾸면 같은 체크인이 11월 1일로 옮겨 간다(저장된 UTC는 그대로)
	if r := a.call(c, "PATCH", "/api/me", map[string]any{"timezone": "America/New_York"}); r.status != http.StatusOK {
		t.Fatalf("시간대 바꾸기: %d %s", r.status, r.body)
	}
	if got := day("2026-11-01"); len(got) != 1 || got[0] != nearMidnight {
		t.Errorf("뉴욕: 11월 1일에 있어야 한다: %v", got)
	}
	if got := day("2026-11-02"); len(got) != 0 {
		t.Errorf("뉴욕: 11월 2일에는 없어야 한다: %v", got)
	}

	// 2026-11-01은 뉴욕 서머타임이 끝나는 날(25시간): 0시 EDT = 04:00Z ~ 다음 날 0시 EST = 05:00Z
	before := post("2026-11-01T03:59:00Z") // 10월 31일 23:59 EDT
	first := post("2026-11-01T04:00:00Z")  // 11월 1일 00:00 EDT
	last := post("2026-11-02T04:59:00Z")   // 11월 1일 23:59 EST (서머타임이 끝나 한 시간 늦게 자정)
	after := post("2026-11-02T05:00:00Z")  // 11월 2일 00:00 EST
	got := strings.Join(day("2026-11-01"), " ")
	for _, id := range []string{first, nearMidnight, last} {
		if !strings.Contains(got, id) {
			t.Errorf("11월 1일(25시간)에 있어야 한다: %s", id)
		}
	}
	for _, id := range []string{before, after} {
		if strings.Contains(got, id) {
			t.Errorf("11월 1일에 없어야 한다: %s", id)
		}
	}
	// 시간순(오름차순)
	if ids := day("2026-11-01"); len(ids) != 3 || ids[0] != first || ids[2] != last {
		t.Errorf("시간순이어야 한다: %v", ids)
	}
}

// 체크인을 저장 · 고치기 · 지우면 통계 캐시 버전이 오른다 (DATA_MODEL 4.4절).
func TestCheckinBumpsStatsVersion(t *testing.T) {
	a := newApp(t)
	c := a.signedIn(testEmail)
	ids := a.emotionIDs(c)
	me := decode[struct{ ID string }](t, a.call(c, "GET", "/api/me", nil))
	key := "stats:ver:" + me.ID
	ctx := context.Background()

	id := decode[apiCheckin](t, a.call(c, "POST", "/api/checkins", map[string]any{"emotions": []any{em(ids["불안"], 50)}})).ID
	a.call(c, "PUT", "/api/checkins/"+id, map[string]any{"emotions": []any{em(ids["불안"], 40)}})
	a.call(c, "DELETE", "/api/checkins/"+id, nil)
	// 실패한 요청은 올리지 않는다
	a.call(c, "POST", "/api/checkins", map[string]any{"emotions": []any{}})

	if v, err := a.rdb.Get(ctx, key).Int(); err != nil || v != 3 {
		t.Errorf("통계 버전: %d %v, 3이어야 한다", v, err)
	}
	// TTL 없는 키를 만들지 않는다 (CLAUDE.md 코드 규칙)
	if ttl := a.rdb.TTL(ctx, key).Val(); ttl <= 0 || ttl > 30*24*time.Hour {
		t.Errorf("통계 버전 TTL: %v", ttl)
	}
}

func mustLoc(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatal(err)
	}
	return loc
}
