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
	"github.com/ghts/lib"
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestMySQL_접속정보(t *testing.T) {
	아이디, 암호, DB명 := fMySQL_접속정보()

	lib.F테스트_다름(t, strings.TrimSpace(아이디), "")
	lib.F테스트_다름(t, strings.TrimSpace(암호), "")
	lib.F테스트_다름(t, strings.TrimSpace(DB명), "")
}

func TestMySQL_DB(t *testing.T) {
	db, 에러 := fMySQL_DB()

	lib.F테스트_에러없음(t, 에러)
	lib.F테스트_에러없음(t, db.Ping())

	db.Close()
}

func f실시간_데이터_저장_MySQL_테스트_도우미(t *testing.T, 테이블명 string, 종목코드 string, 값_생성함수 func() interface{}) {
	버퍼 := new(bytes.Buffer)
	버퍼.WriteString("SELECT COUNT(*) FROM " + 테이블명 + " ")
	버퍼.WriteString("WHERE Code = ?")
	원래_수량, 에러 := f정수값_질의(버퍼.String(), 종목코드)
	lib.F테스트_에러없음(t, 에러)

	const 반복_횟수 = 10
	ch대기열 := make(chan lib.I소켓_메시지, 반복_횟수 + 1)

	for i:=0 ; i< 반복_횟수; i++ {
		값 := 값_생성함수()

		소켓_메시지, 에러 := lib.New소켓_메시지(lib.CBOR, 값)
		lib.F테스트_에러없음(t, 에러)

		ch대기열 <- 소켓_메시지
	}

	에러 = fNH_실시간_데이터_저장_MySQL(ch대기열)
	lib.F테스트_에러없음(t, 에러)

	추가_후_수량, 에러 := f정수값_질의(버퍼.String(), 종목코드)
	lib.F테스트_에러없음(t, 에러)

	lib.F테스트_같음(t, 추가_후_수량, 원래_수량 + 반복_횟수)
}

func Test실시간_데이터_저장_MySQL_호가잔량(t *testing.T) {
	값_생성함수 := func() interface{} {
		값 := new(lib.NH호가_잔량)
		값.M종목코드 = 테스트용_종목코드
		값.M시각 = time.Now()
		값.M매도_호가_모음 = lib.F임의_범위_이내_정수64값_모음(10, 1000, 10000000)
		값.M매도_잔량_모음 = lib.F임의_범위_이내_정수64값_모음(10, 1000, 10000000)
		값.M매수_호가_모음 = lib.F임의_범위_이내_정수64값_모음(10, 1000, 10000000)
		값.M매수_잔량_모음 = lib.F임의_범위_이내_정수64값_모음(10, 1000, 10000000)
		값.M누적_거래량 = lib.F임의_양의_정수64값()

		return 값
	}

	f실시간_데이터_저장_MySQL_테스트_도우미(t, "OfferBid", 테스트용_종목코드, 값_생성함수)
}

func Test실시간_데이터_저장_MySQL_시간외_호가잔량(t *testing.T) {
	값_생성함수 := func() interface{} {
		값 := new(lib.NH시간외_호가잔량)
		값.M종목코드 = 테스트용_종목코드
		값.M시각 = time.Now()
		값.M총_매도호가_잔량 = lib.F임의_범위_이내_정수64값(1000, 10000000)
		값.M총_매수호가_잔량 = lib.F임의_범위_이내_정수64값(1000, 10000000)

		return 값
	}

	f실시간_데이터_저장_MySQL_테스트_도우미(t, "OffTimeOfferBid", 테스트용_종목코드, 값_생성함수)
}

func Test실시간_데이터_저장_MySQL_예상_호가잔량(t *testing.T) {
	값_생성함수 := func() interface{} {
		값 := new(lib.NH예상_호가잔량)
		값.M종목코드 = 테스트용_종목코드
		값.M시각 = time.Now()
		값.M동시호가_구분 = uint8(lib.F임의_양의_정수8값())
		값.M예상_체결가 = lib.F임의_범위_이내_정수64값(1000, 10000000)
		값.M예상_등락부호 = uint8(lib.F임의_양의_정수8값())
		값.M예상_등락폭 = lib.F임의_범위_이내_정수64값(1000, 10000000)
		값.M예상_등락율 = lib.F임의_범위_이내_실수64값(0, 100)
		값.M예상_체결수량 = lib.F임의_범위_이내_정수64값(0, 1000000000)
		값.M매도_호가 = lib.F임의_범위_이내_정수64값(1000, 10000000)
		값.M매수_호가 = lib.F임의_범위_이내_정수64값(1000, 10000000)
		값.M매도_호가잔량 = lib.F임의_범위_이내_정수64값(0, 100000000)
		값.M매수_호가잔량 = lib.F임의_범위_이내_정수64값(0, 100000000)

		return 값
	}

	f실시간_데이터_저장_MySQL_테스트_도우미(t, "EstimatedOfferBid", 테스트용_종목코드, 값_생성함수)
}

func Test실시간_데이터_저장_MySQL_체결(t *testing.T) {
	값_생성함수 := func() interface{} {
		값 := new(lib.NH체결)
		값.M종목코드 = 테스트용_종목코드
		값.M시각 = time.Now()
		값.M등락부호 = uint8(lib.F임의_양의_정수8값())
		값.M등락폭 = lib.F임의_범위_이내_정수64값(1000, 10000000)
		값.M현재가 = lib.F임의_범위_이내_정수64값(1000, 10000000)
		값.M등락율 = lib.F임의_범위_이내_실수64값(0, 100)
		값.M고가 = lib.F임의_범위_이내_정수64값(1000, 10000000)
		값.M저가 = lib.F임의_범위_이내_정수64값(1000, 10000000)
		값.M매도_호가 = lib.F임의_범위_이내_정수64값(1000, 10000000)
		값.M매수_호가 = lib.F임의_범위_이내_정수64값(1000, 10000000)
		값.M누적_거래량 = lib.F임의_범위_이내_정수64값(0, 1000000000)
		값.M전일대비_거래량_비율 = lib.F임의_범위_이내_실수64값(0, 100)
		값.M변동_거래량 = lib.F임의_범위_이내_정수64값(0, 1000000000)
		값.M거래_대금_100만 = lib.F임의_범위_이내_정수64값(0, 1000000000)
		값.M시가 = lib.F임의_범위_이내_정수64값(1000, 10000000)
		값.M가중_평균_가격 = lib.F임의_범위_이내_정수64값(1000, 10000000)
		값.M시장구분 = lib.T시장구분(uint8(lib.F임의_양의_정수8값()))

		return 값
	}

	f실시간_데이터_저장_MySQL_테스트_도우미(t, "Deal", 테스트용_종목코드, 값_생성함수)
}

func Test실시간_데이터_저장_MySQL_ETF_NAV(t *testing.T) {
	값_생성함수 := func() interface{} {
		값 := new(lib.NH_ETF_NAV)
		값.M종목코드 = 테스트용_종목코드
		값.M시각 = time.Now()
		값.M등락부호 = uint8(lib.F임의_양의_정수8값())
		값.M등락폭 = lib.F임의_범위_이내_실수64값(1000, 10000000)
		값.M현재가_NAV = lib.F임의_범위_이내_실수64값(1000, 10000000)
		값.M시가_NAV = lib.F임의_범위_이내_실수64값(1000, 10000000)
		값.M고가_NAV = lib.F임의_범위_이내_실수64값(1000, 10000000)
		값.M저가_NAV = lib.F임의_범위_이내_실수64값(1000, 10000000)
		값.M추적오차_부호 = uint8(lib.F임의_양의_정수8값())
		값.M추적오차 = lib.F임의_범위_이내_실수64값(0, 100)
		값.M괴리율_부호 = uint8(lib.F임의_양의_정수8값())
		값.M괴리율 = lib.F임의_범위_이내_실수64값(0, 100)

		return 값
	}

	f실시간_데이터_저장_MySQL_테스트_도우미(t, "ETF_NAV", 테스트용_종목코드, 값_생성함수)
}

func Test실시간_데이터_저장_MySQL_Sector_Index(t *testing.T) {
	값_생성함수 := func() interface{} {
		값 := new(lib.NH업종지수)
		값.M업종코드 = 테스트용_종목코드[:2]
		값.M시각 = time.Now()
		값.M현재값 = lib.F임의_범위_이내_실수64값(50, 10000)
		값.M등락부호 = uint8(lib.F임의_양의_정수8값())
		값.M등락폭 = lib.F임의_범위_이내_실수64값(50, 10000)
		값.M거래량 = lib.F임의_양의_정수64값()
		값.M거래_대금 = lib.F임의_범위_이내_정수64값(0, 1000000000)
		값.M개장값 = lib.F임의_범위_이내_실수64값(50, 10000)
		값.M최고값 = lib.F임의_범위_이내_실수64값(50, 10000)
		값.M최고값_시각 = time.Now()
		값.M최저값 = lib.F임의_범위_이내_실수64값(50, 10000)
		값.M최저값_시각 = time.Now()
		값.M지수_등락율 = lib.F임의_범위_이내_실수64값(-100, 100)
		값.M거래_비중 = lib.F임의_범위_이내_실수64값(0, 100)

		return 값
	}

	f실시간_데이터_저장_MySQL_테스트_도우미(t, "SectorIndex", 테스트용_종목코드[:2], 값_생성함수)
}
