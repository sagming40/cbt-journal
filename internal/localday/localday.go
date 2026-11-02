// Package localday는 "사용자 시간대의 그날"을 UTC 범위로 바꾼다 (DATA_MODEL 3.4절).
// 체크인 · 사고 기록 · 통계가 같은 기준을 써야 해서 기능이 아닌 작은 패키지에 둔다 (ARCHITECTURE 2절).
package localday

import (
	"time"

	"github.com/sagming40/cbt-journal/internal/apierr"

	// 시간대 DB를 실행 파일에 넣는다. Windows에는 OS에 IANA 시간대 DB가 없어서 이게 없으면 LoadLocation이 실패한다.
	_ "time/tzdata"
)

const dateLayout = "2006-01-02"

// Range는 시간대 tz의 그날 0시 ~ 다음 날 0시를 돌려준다. date가 ""이면 now 기준 오늘.
// 돌려주는 날짜는 실제로 쓴 날짜(YYYY-MM-DD)다 — 화면이 "오늘"을 서버와 같은 기준으로 보이게.
// 서머타임이 끝나는 날은 25시간, 시작하는 날은 23시간 — 날짜 계산을 time.Date에 맡겨 저절로 맞춘다.
func Range(date, tz string, now time.Time) (string, time.Time, time.Time, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		// users.timezone은 저장할 때 검사하므로 여기 올 일은 없다. 그래도 화면이 멈추지 않게 UTC로.
		loc = time.UTC
	}
	if date == "" {
		date = now.In(loc).Format(dateLayout)
	}
	d, err := time.ParseInLocation(dateLayout, date, loc)
	if err != nil {
		return "", time.Time{}, time.Time{}, apierr.Field("date", "날짜 형식이 맞지 않아요 (YYYY-MM-DD)")
	}
	from := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, loc)
	to := time.Date(d.Year(), d.Month(), d.Day()+1, 0, 0, 0, 0, loc)
	return date, from, to, nil
}
