/* Copyright (C) 2015년 김운하(UnHa Kim)  unha.kim@kuh.pe.kr

이 파일은 GHTS의 일부입니다.

이 프로그램은 자유 소프트웨어입니다.
소프트웨어의 피양도자는 자유 소프트웨어 재단이 공표한 GNU LGPL 2.1판
규정에 따라 프로그램을 개작하거나 재배포할 수 있습니다.

이 프로그램은 유용하게 사용될 수 있으리라는 희망에서 배포되고 있지만,
특정한 목적에 적합하다거나, 이익을 안겨줄 수 있다는 묵시적인 보증을 포함한
어떠한 형태의 보증도 제공하지 않습니다.
보다 자세한 사항에 대해서는 GNU LGPL 2.1판을 참고하시기 바랍니다.
GNU LGPL 2.1판은 이 프로그램과 함께 제공됩니다.
만약, 이 문서가 누락되어 있다면 자유 소프트웨어 재단으로 문의하시기 바랍니다.
(자유 소프트웨어 재단 : Free Software Foundation, Inc.,
59 Temple Place - Suite 330, Boston, MA 02111-1307, USA)

Copyright (C) 2015년 UnHa Kim (unha.kim@kuh.pe.kr)

This file is part of GHTS.

GHTS is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as published by
the Free Software Foundation, version 2.1 of the License.

GHTS is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with GHTS.  If not, see <http://www.gnu.org/licenses/>. */

package ghts_utils

import (
	공용 "github.com/ghts/ghts_common"

	"bytes"
	"fmt"
	"sort"
	"strconv"
	"time"
)

const P현금_코드 = "cash"

func f키(종목코드 string, 일자 time.Time) string {
	return 종목코드 + "_" + 일자.Format("20060102")
}

type S일일가격정보_저장소 struct {
	맵       map[string]*공용.S일일_가격정보
	키_코드_일자 map[string]([]time.Time)
	키_일자_코드 map[time.Time]([]string)
}

func (s S일일가격정보_저장소) G종목별_일자_모음(종목코드 string) []time.Time {
	일자_모음, ok := s.키_코드_일자[종목코드]
	if !ok {
		일자_모음 = make([]time.Time, 0)
	}

	return 공용.F슬라이스_복사(공용.F정렬_시각(일자_모음), make([]time.Time, 0)).([]time.Time)
}

func (s S일일가격정보_저장소) G종목별_일자_존재(종목코드 string, 일자 time.Time) bool {
	일자_모음 := s.G종목별_일자_모음(종목코드)
	for _, 현존_일자 := range 일자_모음 {
		switch {
		case 현존_일자.Equal(일자):
			return true
		case 현존_일자.After(일자):
			return false
		}
	}
	return false
}

func (s S일일가격정보_저장소) g일자_도우미(일자_모음 []time.Time, 금일 time.Time, 일자_차이 int) (time.Time, error) {
	금일_인덱스 := 0
	for i, 일자 := range 일자_모음 {
		if 일자.Equal(금일) {
			금일_인덱스 = i
			break
		}
	}

	switch {
	case 금일_인덱스 == 0 && 일자_차이 == -1:
		return 일자_모음[0].AddDate(0, 0, -1), nil
	case 일자_차이 < 0 && 금일_인덱스 < 일자_차이*-1,
		일자_차이 >= 0 && 금일_인덱스+일자_차이 >= len(일자_모음):
		return time.Time{}, 공용.F에러_생성("데이터 없음. %v %v", 금일_인덱스, 일자_차이)
	default:
		return 일자_모음[금일_인덱스+일자_차이], nil
	}

	return time.Time{}, 공용.New에러("금일을 찾을 수 없음. %v", 금일.Format(공용.P일자_형식))
}

func (s S일일가격정보_저장소) g종목별_일자(종목코드 string, 기준일 time.Time, 일자_차이 int) (time.Time, error) {
	return s.g일자_도우미(s.G종목별_일자_모음(종목코드), 기준일, 일자_차이)
}

func (s S일일가격정보_저장소) G종목별_전일(종목코드 string, 기준일 time.Time) (time.Time, error) {
	return s.G종목별_과거_일자(종목코드, 기준일, 1)
}

func (s S일일가격정보_저장소) G종목별_과거_일자(종목코드 string, 기준일 time.Time, 일자_차이 int) (time.Time, error) {
	일자_차이 = int(공용.F2절대값(일자_차이))
	return s.g종목별_일자(종목코드, 기준일, 일자_차이*-1)
}

func (s S일일가격정보_저장소) G종목별_명일(종목코드 string, 기준일 time.Time) (time.Time, error) {
	return s.G종목별_미래_일자(종목코드, 기준일, 1)
}

func (s S일일가격정보_저장소) G종목별_미래_일자(종목코드 string, 기준일 time.Time, 일자_차이 int) (time.Time, error) {
	일자_차이 = int(공용.F2절대값(일자_차이))
	return s.g종목별_일자(종목코드, 기준일, 일자_차이)
}

func (s S일일가격정보_저장소) G전종목_일자_모음() []time.Time {
	일자_모음 := make([]time.Time, 0)
	for 일자, _ := range s.키_일자_코드 {
		일자_모음 = append(일자_모음, 일자)
	}

	return 공용.F정렬_시각(일자_모음)
}

func (s S일일가격정보_저장소) G전종목_일자_존재(일자 time.Time) bool {
	일자_모음 := s.G전종목_일자_모음()
	for _, 현존_일자 := range 일자_모음 {
		switch {
		case 현존_일자.Equal(일자):
			return true
		case 현존_일자.After(일자):
			return false
		}
	}
	return false
}

func (s S일일가격정보_저장소) g전종목_일자(기준일 time.Time, 일자_차이 int) (time.Time, error) {
	return s.g일자_도우미(s.G전종목_일자_모음(), 기준일, 일자_차이)
}

func (s S일일가격정보_저장소) G전종목_전일(금일 time.Time) (time.Time, error) {
	return s.G전종목_과거_일자(금일, 1)
}

func (s S일일가격정보_저장소) G전종목_과거_일자(금일 time.Time, 일자_차이 int) (time.Time, error) {
	일자_차이 = int(공용.F2절대값(일자_차이))
	return s.g전종목_일자(금일, 일자_차이*-1)
}

func (s S일일가격정보_저장소) G전종목_명일(금일 time.Time) (time.Time, error) {
	return s.G전종목_미래_일자(금일, 1)
}

func (s S일일가격정보_저장소) G전종목_미래_일자(금일 time.Time, 일자_차이 int) (time.Time, error) {
	일자_차이 = int(공용.F2절대값(일자_차이))
	return s.g전종목_일자(금일, 일자_차이)
}

func (s S일일가격정보_저장소) G일일_가격정보(종목코드 string, 일자 time.Time) *공용.S일일_가격정보 {
	일일_가격정보, ok := s.맵[f키(종목코드, 일자)]
	if !ok {
		return nil
	}

	return 일일_가격정보.G복사본()
}

func (s S일일가격정보_저장소) G전일_가격정보(종목코드 string, 금일 time.Time) *공용.S일일_가격정보 {
	전일 := 금일

	for {
		전일, 에러 := s.G전종목_전일(전일)
		if 에러 != nil {
			return nil
		}

		if 전일_가격정보 := s.G일일_가격정보(종목코드, 전일); 전일_가격정보 != nil {
			return 전일_가격정보
		}
	}

	return nil
}

func (s S일일가격정보_저장소) G명일_가격정보(종목코드 string, 금일 time.Time) *공용.S일일_가격정보 {
	명일 := 금일

	for {
		명일, 에러 := s.G전종목_명일(명일)
		if 에러 != nil {
			return nil
		}

		if 명일_가격정보 := s.G일일_가격정보(종목코드, 명일); 명일_가격정보 != nil {
			return 명일_가격정보
		}
	}

	return nil
}

func (s S일일가격정보_저장소) G단순이동평균(종목코드 string, 기준일 time.Time,
	기간by일 int, 가격종류 uint8) (float64, error) {
	가격정보_일자_모음 := s.G전종목_일자_모음()
	switch {
	case 기간by일 <= 0:
		return 0, 공용.New에러("잘못된 기간입니다. %v", 기간by일)
	case len(가격정보_일자_모음) < 기간by일,
		기준일.Before(s.G전종목_일자_모음()[기간by일]):
		return 0, 공용.New에러("가격정보가 충분하지 않습니다. %v", 기준일.Format(공용.P일자_형식))
	}

	합계 := 0.0
	일자 := 기준일

	for i := 0; i < 기간by일; i++ {
		var 가격정보 *공용.S일일_가격정보

		// 시간을 거슬러 올라가면서 가격정보를 취득함.
		for {
			가격정보 = s.G일일_가격정보(종목코드, 일자)
			if 가격정보 != nil {
				break
			}

			일자 = 일자.AddDate(0, 0, -1) // 개장일이 아니면, 그 이전 데이터를 검색
			if 일자.Before(가격정보_일자_모음[0]) {
				공용.F패닉("가격정보를 찾을 수 없습니다.")
			}
		}

		switch 가격종류 {
		case P조정시가:
			합계 += float64(가격정보.G조정시가())
		case P조정종가:
			합계 += float64(가격정보.M조정종가)
		}

		일자 = 일자.AddDate(0, 0, -1)
	}

	return 합계 / float64(기간by일), nil
}

func (s *S일일가격정보_저장소) S추가(일일_가격정보 *공용.S일일_가격정보) {
	일자_모음, ok := s.키_코드_일자[일일_가격정보.M종목코드]
	if !ok {
		일자_모음 = make([]time.Time, 0)
	}
	일자_모음 = append(일자_모음, 일일_가격정보.M일자)
	s.키_코드_일자[일일_가격정보.M종목코드] = 일자_모음

	코드_모음, ok := s.키_일자_코드[일일_가격정보.M일자]
	if !ok {
		코드_모음 = make([]string, 0)
	}
	코드_모음 = append(코드_모음, 일일_가격정보.M종목코드)
	s.키_일자_코드[일일_가격정보.M일자] = 코드_모음

	s.맵[일일_가격정보.G키()] = 일일_가격정보
}

func New일일가격정보_저장소() *S일일가격정보_저장소 {
	s := new(S일일가격정보_저장소)
	s.맵 = make(map[string]*공용.S일일_가격정보)
	s.키_코드_일자 = make(map[string]([]time.Time))
	s.키_일자_코드 = make(map[time.Time]([]string))

	return s
}

type S포트폴리오 struct {
	포트폴리오_맵       map[string]*S포트폴리오_구성요소
	포트폴리오_코드_일자_키 map[string](map[time.Time]공용.S비어있음)
	포트폴리오_일자_코드_키 map[time.Time](map[string]공용.S비어있음)
	거래내역_모음       []*S거래내역
}

func (s S포트폴리오) G구성요소(종목코드 string, 일자 time.Time) *S포트폴리오_구성요소 {
	포트폴리오_구성요소, ok := s.포트폴리오_맵[f키(종목코드, 일자)]
	if !ok {
		return nil
	}

	return 포트폴리오_구성요소.G복제본()
}

func (s S포트폴리오) G수량(종목코드 string, 일자 time.Time) int64 {
	포트폴리오_구성요소, ok := s.포트폴리오_맵[f키(종목코드, 일자)]
	if !ok {
		return 0
	}

	return 포트폴리오_구성요소.G수량()
}

func (s S포트폴리오) G일자별_코드_모음(일자 time.Time) []string {
	코드_맵, ok := s.포트폴리오_일자_코드_키[공용.F2일자(일자)]
	if !ok {
		코드_맵 = make(map[string]공용.S비어있음)
	}

	코드_모음 := make([]string, len(코드_맵))
	인덱스 := 0
	for 종목코드, _ := range 코드_맵 {
		코드_모음[인덱스] = 종목코드
		인덱스++
	}

	sort.Strings(코드_모음)

	return 코드_모음
}

func (s S포트폴리오) G현황_출력(일자 time.Time) {
	코드_모음 := s.G일자별_코드_모음(일자)
	평가액_누적 := int64(0)
	for _, 코드 := range 코드_모음 {
		구성요소 := s.G구성요소(코드, 일자)
		fmt.Printf("%v  %v  %v\n", 일자.Format(공용.P일자_형식), 코드, 구성요소.G평가액())
		평가액_누적 += 구성요소.G평가액()
	}
	fmt.Printf("%v 총 평가액 %v\n", 일자.Format(공용.P일자_형식), 평가액_누적)
}

func (s S포트폴리오) G종목별_일자_모음(종목코드 string) []time.Time {
	일자_맵, ok := s.포트폴리오_코드_일자_키[종목코드]
	if !ok {
		일자_맵 = make(map[time.Time]공용.S비어있음)
	}

	일자_모음 := make([]time.Time, len(일자_맵))
	인덱스 := 0
	for 일자, _ := range 일자_맵 {
		일자_모음[인덱스] = 일자
		인덱스++
	}

	return 공용.F정렬_시각(일자_모음)
}

func (s S포트폴리오) G일자_모음() []time.Time {
	일자_모음 := make([]time.Time, len(s.포트폴리오_일자_코드_키))
	인덱스 := 0
	for 일자, _ := range s.포트폴리오_일자_코드_키 {
		일자_모음[인덱스] = 일자
		인덱스++
	}

	return 공용.F정렬_시각(일자_모음)
}

func (s S포트폴리오) G일자별_평가액(일자 time.Time) int64 {
	일자 = 공용.F2일자(일자)
	종목코드_맵, ok := s.포트폴리오_일자_코드_키[일자]
	if !ok {
		return 0
	}

	합계 := int64(0)
	for 종목코드, _ := range 종목코드_맵 {
		포트폴리오_구성요소, ok := s.포트폴리오_맵[f키(종목코드, 일자)]
		if !ok {
			공용.F패닉("예상하지 못한 경우.")
		}

		합계 += 포트폴리오_구성요소.평가액
	}

	return 합계
}

func (s S포트폴리오) G전종목_거래내역_모음() []*S거래내역 {
	return 공용.F슬라이스_복사(s.거래내역_모음, make([]*S거래내역,0)).([]*S거래내역)
}

func (s S포트폴리오) G종목별_거래내역_모음(종목코드 string) []*S거래내역 {
	거래내역_모음 := make([]*S거래내역, 0)
	for _, 거래내역 := range s.거래내역_모음 {
		if 거래내역.G종목코드() == 종목코드 {
			거래내역_모음 = append(거래내역_모음, 거래내역)
		}
	}

	return 거래내역_모음
}

func (s S포트폴리오) G종목별_미결_거래내역_모음(종목코드 string) []*S거래내역 {
	거래내역_모음 := make([]*S거래내역, 0)

	for _, 거래내역 := range s.거래내역_모음 {
		if 거래내역.G거래완료() ||
			거래내역.G종목코드() != 종목코드 {
			continue
		}

		거래내역_모음 = append(거래내역_모음, 거래내역)
	}

	return 거래내역_모음
}

func (s *S포트폴리오) S전일_복제(기준일 time.Time, 가격정보_저장소 *S일일가격정보_저장소) (에러 error) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case error:
				에러 = r.(error)
			default:
				에러 = 공용.New에러("%", r)
			}
		}
	}()

	공용.F조건부_패닉(!가격정보_저장소.G전종목_일자_존재(기준일), "존재하지 않는 일자. %v", 기준일.Format(공용.P일자_형식))
	전일, 에러 := 가격정보_저장소.G전종목_전일(기준일)
	코드_모음 := s.G일자별_코드_모음(전일)
	공용.F조건부_패닉(코드_모음 == nil, "예상하지 못한 경우")

	for _, 종목코드 := range 코드_모음 {
		전일_조정종가 := int64(0)
		전일_가격정보 := 가격정보_저장소.G일일_가격정보(종목코드, 전일)
		금일_구성요소 := s.G구성요소(종목코드, 전일) // 복제

		switch {
		case 종목코드 == P현금_코드:
			전일_조정종가 = 금일_구성요소.G평가액()
		case 전일_가격정보 != nil:
			전일_조정종가 = 전일_가격정보.M조정종가
		default:
			공용.F패닉("전일종가를 알 수 없어서 평가액을 계산할 수 없습니다.")
		}

		// 필요한 부분 업데이트
		금일_구성요소.일자 = 기준일
		금일_구성요소.평가액 = 금일_구성요소.수량 * 전일_조정종가
		금일_구성요소.키 = f키(종목코드, 기준일)

		s.S추가(금일_구성요소)
	}

	return nil
}

func (s *S포트폴리오) S추가(구성요소 *S포트폴리오_구성요소) {
	s.포트폴리오_맵[구성요소.G키()] = 구성요소

	일자_맵, ok := s.포트폴리오_코드_일자_키[구성요소.G종목코드()]
	if !ok {
		일자_맵 = make(map[time.Time]공용.S비어있음)
	}
	일자_맵[구성요소.G일자()] = 공용.S비어있음{}
	s.포트폴리오_코드_일자_키[구성요소.G종목코드()] = 일자_맵

	코드_맵, ok := s.포트폴리오_일자_코드_키[구성요소.G일자()]
	if !ok {
		코드_맵 = make(map[string]공용.S비어있음)
	}
	코드_맵[구성요소.G종목코드()] = 공용.S비어있음{}
	s.포트폴리오_일자_코드_키[구성요소.G일자()] = 코드_맵
}

func (s *S포트폴리오) S매입(종목코드 string, 일자 time.Time, 수량, 단가, 손절매_단가 int64) {
	포트폴리오_구성요소, ok := s.포트폴리오_맵[f키(종목코드, 일자)]
	if !ok {
		포트폴리오_구성요소 = New포트폴리오_구성요소(종목코드, 일자, 0, 단가)
	}

	포트폴리오_구성요소.s매입(수량, 단가)
	s.S추가(포트폴리오_구성요소)

	if 종목코드 == P현금_코드 {
		return
	}

	현금, ok := s.포트폴리오_맵[f키(P현금_코드, 일자)]
	if !ok {
		공용.New에러("예상하지 못한 경우.")

		일자_모음 := s.G일자_모음()
		for _, 일자 := range 일자_모음 {
			코드_모음 := s.G일자별_코드_모음(일자)
			for _, 코드 := range 코드_모음 {
				구성요소 := s.G구성요소(코드, 일자)
				공용.F체크_포인트(구성요소.String())
			}
		}

		현금 = New현금(일자, 0)
	}

	현금.s매도(1, 수량*단가)
	s.S추가(현금)

	거래내역 := New거래내역(종목코드, 수량, 일자, 단가, 손절매_단가)
	s.거래내역_모음 = append(s.거래내역_모음, 거래내역)
}

func (s *S포트폴리오) S매도(종목코드 string, 일자 time.Time, 수량, 단가 int64) {
	포트폴리오_구성요소, ok := s.포트폴리오_맵[f키(종목코드, 일자)]
	if !ok {
		공용.F체크_포인트()
		포트폴리오_구성요소 = New포트폴리오_구성요소(종목코드, 일자, 0, 단가)
	}

	포트폴리오_구성요소.s매도(수량, 단가)

	if 포트폴리오_구성요소.수량 == 0 {
		delete(s.포트폴리오_맵, 포트폴리오_구성요소.G키())
		delete(s.포트폴리오_일자_코드_키[일자], 종목코드)
		delete(s.포트폴리오_코드_일자_키[종목코드], 일자)
	} else {
		s.S추가(포트폴리오_구성요소)
	}

	if 종목코드 == P현금_코드 {
		return
	}

	현금, ok := s.포트폴리오_맵[f키(P현금_코드, 일자)]
	if !ok {
		공용.New에러("예상하지 못한 경우.")
		현금 = New현금(일자, 0)
	}

	현금.s매입(1, 수량*단가)
	s.S추가(현금)

	미결_거래내역_모음 := s.G종목별_미결_거래내역_모음(종목코드)
	누적_매도수량 := int64(0)

반복문:
	for _, 미결_거래내역 := range 미결_거래내역_모음 {
		잔여_수량 := 수량 - 누적_매도수량
		switch {
		case 미결_거래내역.G수량() > 잔여_수량:
			공용.F패닉("'S거래내역'은 1회에 전량 매도하는 것을 가정하여 설계되었습니다.")
		case 미결_거래내역.G수량() == 잔여_수량:
			미결_거래내역.S매도기록(일자, 단가)
			break 반복문
		case 미결_거래내역.G수량() < 잔여_수량:
			미결_거래내역.S매도기록(일자, 단가)
			누적_매도수량 += 미결_거래내역.G수량()
			continue
		default:
			공용.F패닉("예상하지 못한 경우")
		}
	}
}

func (s *S포트폴리오) S전량_매도(종목코드 string, 일자 time.Time, 단가 int64) {
	수량 := s.G수량(종목코드, 일자)
	s.S매도(종목코드, 일자, 수량, 단가)
}

func (s *S포트폴리오) S종목별_손절매(가격정보_저장소 *S일일가격정보_저장소, 종목코드 string,
	금일 time.Time, 장_개시_종료 uint8) {
	구성요소 := s.G구성요소(종목코드, 금일)
	if 구성요소 == nil {
		return
	}

	금일_가격정보 := 가격정보_저장소.G일일_가격정보(종목코드, 금일)
	if 금일_가격정보 == nil {
		return
	}

	거래내역_모음 := s.G종목별_미결_거래내역_모음(종목코드)
	for _, 거래내역 := range 거래내역_모음 {
		switch {
		case 거래내역.G매입일().After(금일):
			continue
		case 장_개시_종료 == P장_개시:
			if 금일_가격정보.G조정시가() > 거래내역.G손절매_단가() {
				continue
			}

			구성요소.s전량_매도(금일_가격정보.M조정종가)
		case 장_개시_종료 == P장_종료:
			if 금일_가격정보.M조정종가 > 거래내역.G손절매_단가() {
				continue
			}

			명일_가격정보 := 가격정보_저장소.G명일_가격정보(종목코드, 금일)
			if 명일_가격정보 == nil {
				continue
			}

			s.S전량_매도(거래내역.G종목코드(), 금일, 명일_가격정보.G조정시가())
		default:
			공용.F패닉("예상하지 못한 경우.")
		}
	}
}

func (s *S포트폴리오) S전종목_손절매(가격정보_저장소 *S일일가격정보_저장소, 금일 time.Time, 장_개시_종료 uint8) {
	금일 = 공용.F2일자(금일)

	종목코드_모음 := s.G일자별_코드_모음(금일)
	for _, 종목코드 := range 종목코드_모음 {
		s.S종목별_손절매(가격정보_저장소, 종목코드, 금일, 장_개시_종료)
	}
}

func (s *S포트폴리오) S금일_종가_매입(가격정보_저장소 *S일일가격정보_저장소, 종목코드 string, 금일 time.Time, 수량 int64, 손절매_비율 float64) {
	금일_가격정보 :=  가격정보_저장소.G일일_가격정보(종목코드, 금일)
	공용.F조건부_패닉(금일_가격정보 == nil, "금일 가격정보를 찾을 수 없습니다. %v", 금일.Format(공용.P일자_형식))
	매입_단가 := 금일_가격정보.M조정종가
	손절매_단가 := int64(float64(매입_단가) * (100.0 - 손절매_비율) / 100.0)
	s.S매입(종목코드, 금일, 수량, 매입_단가, 손절매_단가)
}

func (s *S포트폴리오) S명일_시가_매입(가격정보_저장소 *S일일가격정보_저장소, 종목코드 string, 금일 time.Time, 수량 int64, 손절매_비율 float64) {
	명일, 에러 := 가격정보_저장소.G종목별_명일(종목코드, 금일)
	공용.F에러_패닉(에러)
	명일_가격정보 :=  가격정보_저장소.G일일_가격정보(종목코드, 명일)
	공용.F조건부_패닉(명일_가격정보 == nil, "명일 가격정보를 찾을 수 없습니다. %v", 명일.Format(공용.P일자_형식))
	매입_단가 := 명일_가격정보.G조정시가()
	손절매_단가 := int64(float64(매입_단가) * (100.0 - 손절매_비율) / 100.0)
	s.S매입(종목코드, 금일, 수량, 매입_단가, 손절매_단가)
}

func (s *S포트폴리오) S금일_종가_매도(가격정보_저장소 *S일일가격정보_저장소, 종목코드 string, 금일 time.Time, 수량 int64) {
	금일_가격정보 :=  가격정보_저장소.G일일_가격정보(종목코드, 금일)
	공용.F조건부_패닉(금일_가격정보 == nil, "금일 가격정보를 찾을 수 없습니다. %v", 금일.Format(공용.P일자_형식))
	매도_단가 := 금일_가격정보.M조정종가
	s.S매도(종목코드, 금일, 수량, 매도_단가)
}

func (s *S포트폴리오) S명일_시가_매도(가격정보_저장소 *S일일가격정보_저장소, 종목코드 string, 금일 time.Time, 수량 int64) {
	명일, 에러 := 가격정보_저장소.G종목별_명일(종목코드, 금일)
	공용.F에러_패닉(에러)
	명일_가격정보 :=  가격정보_저장소.G일일_가격정보(종목코드, 명일)
	공용.F조건부_패닉(명일_가격정보 == nil, "명일 가격정보를 찾을 수 없습니다. %v", 명일.Format(공용.P일자_형식))
	매도_단가 := 명일_가격정보.M조정종가
	s.S매도(종목코드, 금일, 수량, 매도_단가)
}

func New포트폴리오() *S포트폴리오 {
	s := new(S포트폴리오)
	s.포트폴리오_맵 = make(map[string]*S포트폴리오_구성요소)
	s.포트폴리오_코드_일자_키 = make(map[string](map[time.Time]공용.S비어있음))
	s.포트폴리오_일자_코드_키 = make(map[time.Time](map[string]공용.S비어있음))
	s.거래내역_모음 = make([]*S거래내역, 0)

	return s
}

type S포트폴리오_구성요소 struct {
	종목코드 string
	일자   time.Time
	수량   int64
	평가액  int64
	키    string
}

func (s S포트폴리오_구성요소) G종목코드() string  { return s.종목코드 }
func (s S포트폴리오_구성요소) G일자() time.Time { return s.일자 }
func (s S포트폴리오_구성요소) G수량() int64     { return s.수량 }
func (s S포트폴리오_구성요소) G평가액() int64    { return s.평가액 }
func (s S포트폴리오_구성요소) G키() string     { return s.키 }
func (s S포트폴리오_구성요소) G복제본() *S포트폴리오_구성요소 {
	복제본 := new(S포트폴리오_구성요소)
	복제본.종목코드 = s.종목코드
	복제본.일자 = 공용.F2일자(s.일자)
	복제본.수량 = s.수량
	복제본.평가액 = s.평가액
	복제본.키 = s.종목코드 + "_" + s.일자.Format("20060102")

	return 복제본
}

func (s *S포트폴리오_구성요소) s매입(수량, 단가 int64) {
	switch s.종목코드 {
	case P현금_코드:
		if 수량 != 1 {
			공용.F패닉("현금은 수량을 1로 해야 합니다. %v", 수량)
		}

		s.수량 = 1
		s.평가액 += 단가
	default:
		s.수량 += 수량
		s.평가액 += 수량 * 단가
	}
}

func (s *S포트폴리오_구성요소) s매도(수량, 단가 int64) {
	switch s.종목코드 {
	case P현금_코드:
		if 수량 != 1 {
			공용.F패닉("현금은 수량을 1로 해야 합니다. %v", 수량)
		}

		s.수량 = 1
		s.평가액 -= 단가
	default:
		s.수량 -= 수량
		s.평가액 -= 수량 * 단가
	}
}

func (s *S포트폴리오_구성요소) s전량_매도(단가 int64) {
	s.s매도(s.수량, 단가)
}

func (s S포트폴리오_구성요소) String() string {
	버퍼 := new(bytes.Buffer)
	버퍼.WriteString(s.키)
	버퍼.WriteString(" : ")
	버퍼.WriteString(s.종목코드)
	버퍼.WriteString(", ")
	버퍼.WriteString(s.일자.Format(공용.P일자_형식))
	버퍼.WriteString(", ")
	버퍼.WriteString(strconv.FormatInt(s.수량, 10))
	버퍼.WriteString(", ")
	버퍼.WriteString(strconv.FormatInt(s.평가액, 10))

	return 버퍼.String()
}

func New포트폴리오_구성요소(종목코드 string, 일자 time.Time, 수량, 평가액_기준가 int64) *S포트폴리오_구성요소 {
	s := new(S포트폴리오_구성요소)
	s.종목코드 = 종목코드
	s.일자 = 공용.F2일자(일자)
	s.수량 = 수량
	s.평가액 = 수량 * 평가액_기준가
	s.키 = 종목코드 + "_" + s.일자.Format("20060102")

	return s
}

func New현금(일자 time.Time, 금액 int64) *S포트폴리오_구성요소 {
	s := new(S포트폴리오_구성요소)
	s.종목코드 = P현금_코드
	s.일자 = 공용.F2일자(일자)
	s.수량 = 1
	s.평가액 = 금액
	s.키 = P현금_코드 + "_" + s.일자.Format("20060102")

	return s
}

// 문제를 단순화 하기 위해서 1회 매입한 수량은 1회에 모두 매도하는 경우로 한정.
type S거래내역 struct {
	종목코드   string
	수량     int64
	매입일    time.Time
	매입_단가  int64
	매도일    time.Time
	매도_단가  int64
	손절매_단가 int64 // 매입 시에만 설정.
}

func (s S거래내역) G종목코드() string   { return s.종목코드 }
func (s S거래내역) G수량() int64      { return s.수량 }
func (s S거래내역) G매입일() time.Time { return s.매입일 }
func (s S거래내역) G매입_단가() int64   { return s.매입_단가 }
func (s S거래내역) G매도일() time.Time { return s.매도일 }
func (s S거래내역) G매도_단가() int64   { return s.매도_단가 }
func (s S거래내역) G손절매_단가() int64  { return s.손절매_단가 }
func (s S거래내역) G거래완료() bool     { return !s.매도일.Equal(time.Time{}) || s.매도_단가 != 0 }
func (s *S거래내역) S매도기록(매도일 time.Time, 매도_단가 int64) {
	switch {
	case 매도일.Before(s.매입일):
		공용.F패닉("잘못된 일자. %v %v", s.매입일.Format(공용.P일자_형식), s.매도일.Format(공용.P일자_형식))
	case 매도_단가 <= 0:
		공용.F패닉("잘못된 매도 단가. %v", s.매도_단가)
	}

	s.매도일 = 매도일
	s.매도_단가 = 매도_단가
}

func New거래내역(종목코드 string, 수량 int64, 매입일 time.Time,
	매입_단가, 손절매_단가 int64) *S거래내역 {
	s := new(S거래내역)
	s.종목코드 = 종목코드
	s.수량 = 수량
	s.매입일 = 매입일
	s.매입_단가 = 매입_단가
	s.손절매_단가 = 손절매_단가

	return s
}

func F정수64_비율(정수값 int64, 비율 float64) int64 {
	return int64(float64(정수값) * 비율)
}