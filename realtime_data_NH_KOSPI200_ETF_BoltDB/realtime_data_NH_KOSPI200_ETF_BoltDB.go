/* Copyright (C) 2015-2016 김운하(UnHa Kim)  unha.kim@kuh.pe.kr

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

package main

import (
	nh "github.com/ghts/api_helper_nh"
	"github.com/ghts/lib"
	"time"
	"strings"
	"os"
	"runtime"
)

func main() {
	// TODO :  시작할 때 시간 동기화

	var 에러 error

	defer lib.F에러패닉_처리(lib.S에러패닉_처리{
		M에러: &에러,
		M함수 : func() { lib.F에러_출력(에러) }})

	_, 파일명, _, ok := runtime.Caller(0)
	lib.F조건부_패닉(!ok, "디렉토리를 찾을 수 없습니다.")

	작업_디렉토리, 에러 := lib.F디렉토리명(파일명)
	lib.F에러2패닉(에러)
	lib.F에러2패닉(os.Chdir(작업_디렉토리))

	lib.F에러2패닉(nh.F접속_NH())

	lib.F문자열_출력("접속됨")

	종목_모음 := []*lib.S종목{
		lib.New종목("069500", "KODEX 200", lib.P시장구분_ETF),
		lib.New종목("114800", "KODEX 인버스", lib.P시장구분_ETF),
		lib.New종목("122630", "KODEX 레버리지", lib.P시장구분_ETF),
		lib.New종목("252670", "KODEX 200 선물인버스2X", lib.P시장구분_ETF),
		lib.New종목("069660", "KOSEF 200", lib.P시장구분_ETF),
		lib.New종목("152280", "KOSEF 200선물", lib.P시장구분_ETF),
		lib.New종목("253250", "KOSEF 200 선물레버리지", lib.P시장구분_ETF),
		lib.New종목("253240", "KOSEF 200 선물인버스", lib.P시장구분_ETF),
		lib.New종목("253230", "KOSEF 200 선물인버스2X", lib.P시장구분_ETF),
		lib.New종목("102110", "TIGER 200", lib.P시장구분_ETF),
		lib.New종목("252710", "TIGER 200 선물인버스2X", lib.P시장구분_ETF),
		lib.New종목("105190", "KINDEX 200", lib.P시장구분_ETF),
		lib.New종목("108590", "TREX 200", lib.P시장구분_ETF),
		lib.New종목("148020", "KBSTAR 200", lib.P시장구분_ETF),
		lib.New종목("252400", "KBSTAR 200 선물레버리지", lib.P시장구분_ETF),
		lib.New종목("252410", "KBSTAR 200 선물인버스", lib.P시장구분_ETF),
		lib.New종목("252420", "KBSTAR 200 선물인버스2X", lib.P시장구분_ETF),
		lib.New종목("152100", "ARIRANG 200", lib.P시장구분_ETF),
		lib.New종목("253150", "ARIRANG 200 선물레버리지", lib.P시장구분_ETF),
		lib.New종목("253160", "ARIRANG 200 선물인버스2X", lib.P시장구분_ETF)}

	// 종목 모음 내용 검사.
	if len(lib.F중복_종목_제거(종목_모음)) != len(종목_모음) {
		lib.F패닉("중복 종목이 존재합니다.")
	}

	for _, 종목 := range 종목_모음 {
		검색된_종목, 에러 := lib.F종목by코드(종목.G코드())
		lib.F에러2패닉(에러)
		lib.F조건부_패닉(strings.Replace(종목.G이름(), " ", "", -1) != strings.Replace(검색된_종목.G이름(), " ", "", -1),
			"잘못된 종목 이름. %v %v", 종목.G이름(), 검색된_종목.G이름())
		lib.F조건부_패닉(종목.G시장구분() != 검색된_종목.G시장구분(),
			"잘못된 시장 구분. %v %v", 종목.G시장구분(), 검색된_종목.G시장구분())
	}

	lib.F문자열_출력("종목 %v개 설정 완료", len(종목_모음))

	종목코드_모음 := lib.F2종목코드_모음(종목_모음)

	nh.F접속유지()	// 공유기 사용 시 접속 끊기는 것 방지.

	db, 에러 := f실시간_데이터_수집_초기화(종목코드_모음)
	lib.F에러2패닉(에러)

	lib.F문자열_출력("실시간 데이터 수집 시작")

	버킷ID_호가_잔량 := []byte(nh.P버킷ID_NH호가_잔량)
	버킷ID_시간외_호가_잔량 := []byte(nh.P버킷ID_NH시간외_호가잔량)
	버킷ID_예상_호가_잔량 := []byte(nh.P버킷ID_NH예상_호가잔량)
	버킷ID_체결 := []byte(nh.P버킷ID_NH체결)
	버킷ID_ETF_NAV := []byte(nh.P버킷ID_NH_ETF_NAV)
	버킷ID_업종지수 := []byte(nh.P버킷ID_NH업종지수)

	금일_문자열 := time.Now().Format(lib.P일자_형식)

	저장수량_체크 := time.NewTicker(lib.P1분)
	일자바뀜_체크 := time.NewTicker(lib.P10초)

	if lib.F테스트_모드_실행_중() {
		저장수량_체크 = time.NewTicker(lib.P10초)
	}

	defer func() {
		저장수량_체크.Stop()
		일자바뀜_체크.Stop()
		nh.F실시간_데이터_해지_NH_ETF(종목코드_모음)
	}()

	for {
		select {
		case <-저장수량_체크.C:
			저장_수량 := db.G수량in버킷(버킷ID_호가_잔량) +
				db.G수량in버킷(버킷ID_시간외_호가_잔량) +
				db.G수량in버킷(버킷ID_예상_호가_잔량) +
				db.G수량in버킷(버킷ID_체결) +
				db.G수량in버킷(버킷ID_ETF_NAV) +
				db.G수량in버킷(버킷ID_업종지수)

			lib.F문자열_출력("%s : %v", time.Now().Format(lib.P간략한_시간_형식), 저장_수량)

			if lib.F테스트_모드_실행_중() {
				return    // 테스트 할 때는 반복할 필요없음.
			}
		case <-일자바뀜_체크.C:
			if 금일_문자열 == time.Now().Format(lib.P일자_형식) {
				continue
			}

			lib.F공통_종료_채널_닫은_후_재설정()
			db, 에러 = f실시간_데이터_수집_초기화(종목코드_모음)
			lib.F에러2패닉(에러)

			금일_문자열 = time.Now().Format(lib.P일자_형식)
			lib.F문자열_출력("실시간 데이터 수집 재시작. %s", 금일_문자열)
		}
	}
}

func f실시간_데이터_수집_초기화(종목코드_모음 []string) (db lib.I데이터베이스_Bolt, 에러 error) {
	defer lib.F에러패닉_처리(lib.S에러패닉_처리{
		M에러: &에러,
		M함수 : func() { db = nil }})

	금일_문자열 := time.Now().Format(lib.P일자_형식)
	파일명 := "RealTimeData_NH_" + 금일_문자열 + ".dat"

	return nh.F실시간_데이터_수집_NH_ETF(파일명, 종목코드_모음)
}