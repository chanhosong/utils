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
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"sync"
)

// 테스트할 때는 도우미를 MOCK-UP(모의 데이터)으로 교체함.
//var 일일가격정보_불러오기_도우미 = 일일가격정보_불러오기_도우미_야후_한국

var 종목별_일일가격정보_질의_도우미 = 종목별_일일가격정보_질의_도우미_야후
var M가격정보_맵_파일_잠금 = new(sync.Mutex)
const P가격정보_맵_파일명 = "yahoo_daily_price.dat"

func F전종목_일일가격정보_확보_야후() (map[string]([]*공용.S일일_가격정보), map[string]*공용.S에러내역) {
	가격정보_맵 := make(map[string]([]*공용.S일일_가격정보))
	에러내역_맵 := make(map[string]*공용.S에러내역)
	종목모음_질의대상, 에러 := 공용.F종목모음_전체()
	공용.F에러_패닉(에러)

	종목모음_에러발생 := make([]*공용.S종목, 0)

	for 반복횟수 := 0; 반복횟수 < 3; 반복횟수++ {
		for _, 종목 := range 종목모음_질의대상 {
			if 종목.G시장구분() != 공용.P시장구분_ETF &&
				종목.G시장구분() != 공용.P시장구분_코스피 {
				continue
			}

			공용.F문자열_출력("종목코드 %v, 종목명칭 %v", 종목.G코드(), 종목.G이름())

			종목별_가격정보_모음, 에러내역 := F종목별_일일가격정보_질의(종목)
			if 에러내역 != nil {
				에러내역_맵[종목.G코드()] =에러내역
				공용.F문자열_출력("%v", 에러내역)
				종목모음_에러발생 = append(종목모음_에러발생, 종목)

				time.Sleep(공용.P30초)
				continue
			}

			가격정보_맵[종목.G코드()] = 종목별_가격정보_모음
			time.Sleep(공용.P10초)
		}

		// 오류가 발생한 종목에 대해서 최대 3회 재시도
		종목모음_질의대상 = 종목모음_에러발생
		종목모음_에러발생 = make([]*공용.S종목, 0)
	}

	// 반복 재시도 이후에 남은 오류 기록 보여주기
	fmt.Println("")
	fmt.Println("-----------------------------------")
	fmt.Println("일일가격정보 오류 발생 기록.")
	fmt.Println("-----------------------------------")
	fmt.Println("")

	for _, 종목 := range 종목모음_에러발생 {
		에러내역 := 에러내역_맵[종목.G코드()]

		공용.F문자열_출력("%v : %v %v", 종목, 에러내역.G에러_코드(), 에러내역.G에러_메시지())
	}

	공용.F파일에_값_저장(가격정보_맵, P가격정보_맵_파일명, M가격정보_맵_파일_잠금)

	return 가격정보_맵, 에러내역_맵
}

func F종목별_일일가격정보_질의(종목 *공용.S종목) ([]*공용.S일일_가격정보, *공용.S에러내역) {
	시작일 := 공용.F2포맷된_일자_단순형("2006-01-02", "1900-01-01")
	종료일 := 공용.F2포맷된_일자_단순형("2006-01-02", time.Now().Format("2006-01-02"))

	일일가격정보_문자열_모음, 에러내역 := 종목별_일일가격정보_질의_도우미(종목, 시작일, 종료일)
	if 에러내역 != nil {
		return nil, 에러내역
	}

	종목별_일일가격정보_맵, 에러내역 := 종목별_일일가격정보_모음_생성_도우미(종목, 일일가격정보_문자열_모음)
	if 에러내역 != nil {
		return nil, 에러내역
	}

	return 종목별_일일가격정보_맵, nil
}

func 종목별_일일가격정보_질의_도우미_야후(종목 *공용.S종목, 시작일 time.Time, 종료일 time.Time) ([][]string, *공용.S에러내역) {
	응답, 에러 := http.Get(야후_가격정보_질의_URL생성(종목, 시작일, 종료일))
	//응답, 에러 := http.Get(`http://ichart.finance.yahoo.com/table.csv?g=d&f=2014&e=12&c=2014&b=10&a=7&d=7&s=069500.KS`)

	switch {
	case 에러 != nil:
		공용.F체크_포인트()
		return nil, 공용.New에러내역("HTTP요청 에러", 에러.Error())
	case 응답.Body == nil:
		공용.F체크_포인트()
		return nil, 공용.New에러내역("nil 응답", "응답.Body가 nil입니다.")
	}

	defer 응답.Body.Close() // panic이 발생하면 하지 말 것.

	레코드_모음, 에러 := csv.NewReader(응답.Body).ReadAll()

	switch {
	case 에러 != nil:
		return nil, 공용.New에러내역("CSV읽기 에러", 에러.Error() + "\n")
	case len(레코드_모음) < 1:
		// 가격데이터가 없거나, CSV로 해석불가능하면 무시
		return nil, 공용.New에러내역("CSV읽기 에러", "CSV 레코드 수량이 0개")
	}

	공용.F변수값_확인(응답, 응답.Body, 레코드_모음)

	// 첫째 줄(레코드_모음[0])은  제목이라서 제거한다.
	레코드_모음 = 레코드_모음[1:]

	return 레코드_모음, nil
}

func 야후_가격정보_질의_URL생성(종목 *공용.S종목, 시작일 time.Time, 종료일 time.Time) string {
	var 시장구분 string

	switch 종목.G시장구분() {
	case 공용.P시장구분_코스피, 공용.P시장구분_ETF:
		시장구분 = ".KS" // 유가증권 시장
	case 공용.P시장구분_코스닥:
		시장구분 = ".KQ" // 코스닥 시장
	}

	종목코드_문자열 := strings.TrimPrefix(종목.G코드(), "A") + 시장구분
	종목코드_문자열 = url.QueryEscape(종목코드_문자열)

	기본URL주소 := "http://ichart.yahoo.com/table.csv?s="

	연도_시작, 월_시작, 일_시작 := 야후_연월일_문자열(시작일)
	연도_종료, 월_종료, 일_종료 := 야후_연월일_문자열(종료일)

	일일가격정보형식 := "g=d"
	//주간데이터형식 := "g=w"
	//월간데이터형식 := "g=m"

	CSV형식 := "ignore=.csv"

	var URL요청문자 bytes.Buffer
	URL요청문자.WriteString(기본URL주소)
	URL요청문자.WriteString(종목코드_문자열)
	URL요청문자.WriteString("&a=" + 월_시작)
	URL요청문자.WriteString("&b=" + 일_시작)
	URL요청문자.WriteString("&c=" + 연도_시작)
	URL요청문자.WriteString("&d=" + 월_종료)
	URL요청문자.WriteString("&e=" + 일_종료)
	URL요청문자.WriteString("&f=" + 연도_종료)
	URL요청문자.WriteString("&" + 일일가격정보형식)
	URL요청문자.WriteString("&" + CSV형식)

	return URL요청문자.String()
}

func 야후_연월일_문자열(일자 time.Time) (연도 string, 월 string, 일 string) {
	// Yahoo! Finanace API규격상 월(Month)은 1을 빼주어야 한다고 함. Java도 이런 식이라서 엄청 헷갈렸는 데.
	// (참고자료 : https://code.google.com/p/yahoo-finance-managed/wiki/csvHistQuotesDownload)
	연도 = strconv.Itoa(일자.Year())
	월 = strconv.Itoa(int(일자.Month()) - 1)
	일 = strconv.Itoa(일자.Day())

	return 연도, 월, 일
}

func 종목별_일일가격정보_모음_생성_도우미(종목 *공용.S종목, 레코드_모음 [][]string) (
	종목별_가격정보_모음 []*공용.S일일_가격정보, 에러내역 *공용.S에러내역) {
	defer func() {
		if r := recover(); r != nil {
			종목별_가격정보_모음 = nil
			에러내역 = 공용.New에러내역("일일가격정보 생성 에러", 공용.F포맷된_문자열("%", r))
		}
	}()

	종목별_가격정보_모음 = make([]*공용.S일일_가격정보, 0)
	const 날짜형식 = "2006-01-02"

	for _, 레코드 := range 레코드_모음 {
		일자 := 공용.F2포맷된_일자_단순형(날짜형식, 레코드[0])
		시가 := int64(공용.F2실수_단순형(레코드[1]))
		고가 := int64(공용.F2실수_단순형(레코드[2]))
		저가 := int64(공용.F2실수_단순형(레코드[3]))
		종가 := int64(공용.F2실수_단순형(레코드[4]))
		거래량 := int64(공용.F2실수_단순형(레코드[5]))
		조정종가 := int64(공용.F2실수_단순형(레코드[6]))

		일일가격정보 := new(공용.S일일_가격정보)
		일일가격정보.M일자 = 일자
		일일가격정보.M종목코드 = 종목.G코드()
		일일가격정보.M고가 = 고가
		일일가격정보.M저가 = 저가
		일일가격정보.M시가 = 시가
		일일가격정보.M종가 = 종가
		일일가격정보.M조정종가 = 조정종가
		일일가격정보.M거래량 = 거래량

		종목별_가격정보_모음 = append(종목별_가격정보_모음, 일일가격정보)
	}

	return 종목별_가격정보_모음, nil
}