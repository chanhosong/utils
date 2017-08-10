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
	"reflect"
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

func f실시간_데이터_저장_MySQL_테스트_도우미(t *testing.T,
	테이블명 string, 종목코드 string, 값_생성함수 func() interface{}) (값_모음 []interface{}) {

	버퍼 := new(bytes.Buffer)
	버퍼.WriteString("DELETE FROM " + 테이블명 + " ")
	버퍼.WriteString("WHERE Code = ?")
	lib.F테스트_에러없음(t, fTX실행(버퍼.String(), 종목코드))

	const 반복_횟수 = 10
	값_모음 = make([]interface{}, 반복_횟수)
	ch대기열 := make(chan lib.I소켓_메시지, 반복_횟수)

	for i := 0; i < 반복_횟수; i++ {
		값 := 값_생성함수()
		값_모음[i] = 값

		소켓_메시지, 에러 := lib.New소켓_메시지(lib.CBOR, 값)
		lib.F테스트_에러없음(t, 에러)

		ch대기열 <- 소켓_메시지
	}

	에러 := fNH_실시간_데이터_저장_MySQL(ch대기열)
	lib.F테스트_에러없음(t, 에러)

	버퍼 = new(bytes.Buffer)
	버퍼.WriteString("SELECT COUNT(*) FROM " + 테이블명 + " ")
	버퍼.WriteString("WHERE Code = ?")
	수량, 에러 := f질의_정수값(버퍼.String(), 종목코드)
	lib.F테스트_에러없음(t, 에러)

	lib.F테스트_같음(t, 수량, 반복_횟수)

	return 값_모음
}

func Test실시간_데이터_저장_MySQL_호가잔량(t *testing.T) {
	const 테이블명 = "OfferBid"

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

	값_모음 := f실시간_데이터_저장_MySQL_테스트_도우미(t, 테이블명, 테스트용_종목코드, 값_생성함수)

	버퍼 := new(bytes.Buffer)
	버퍼.WriteString("SELECT * FROM " + 테이블명 + " ")
	버퍼.WriteString("WHERE Code = ?")

	레코드_모음, 에러 := f질의_Rows(버퍼.String(), 테스트용_종목코드)
	lib.F테스트_에러없음(t, 에러)

	defer 레코드_모음.Close()
	for 레코드_모음.Next() {
		var id int
		var 코드 string
		var 시각 time.Time
		var 매도_호가_1, 매도_호가_2, 매도_호가_3, 매도_호가_4, 매도_호가_5 string
		var 매도_호가_6, 매도_호가_7, 매도_호가_8, 매도_호가_9, 매도_호가_10 string
		var 매수_호가_1, 매수_호가_2, 매수_호가_3, 매수_호가_4, 매수_호가_5 string
		var 매수_호가_6, 매수_호가_7, 매수_호가_8, 매수_호가_9, 매수_호가_10 string
		var 매도_호가_잔량_1, 매도_호가_잔량_2, 매도_호가_잔량_3, 매도_호가_잔량_4, 매도_호가_잔량_5 int64
		var 매도_호가_잔량_6, 매도_호가_잔량_7, 매도_호가_잔량_8, 매도_호가_잔량_9, 매도_호가_잔량_10 int64
		var 매수_호가_잔량_1, 매수_호가_잔량_2, 매수_호가_잔량_3, 매수_호가_잔량_4, 매수_호가_잔량_5 int64
		var 매수_호가_잔량_6, 매수_호가_잔량_7, 매수_호가_잔량_8, 매수_호가_잔량_9, 매수_호가_잔량_10 int64
		var 누적_거래량 int64

		에러 = 레코드_모음.Scan(&id, &코드, &시각,
			&매도_호가_1, &매수_호가_1, &매도_호가_잔량_1, &매수_호가_잔량_1,
			&매도_호가_2, &매수_호가_2, &매도_호가_잔량_2, &매수_호가_잔량_2,
			&매도_호가_3, &매수_호가_3, &매도_호가_잔량_3, &매수_호가_잔량_3,
			&매도_호가_4, &매수_호가_4, &매도_호가_잔량_4, &매수_호가_잔량_4,
			&매도_호가_5, &매수_호가_5, &매도_호가_잔량_5, &매수_호가_잔량_5,
			&매도_호가_6, &매수_호가_6, &매도_호가_잔량_6, &매수_호가_잔량_6,
			&매도_호가_7, &매수_호가_7, &매도_호가_잔량_7, &매수_호가_잔량_7,
			&매도_호가_8, &매수_호가_8, &매도_호가_잔량_8, &매수_호가_잔량_8,
			&매도_호가_9, &매수_호가_9, &매도_호가_잔량_9, &매수_호가_잔량_9,
			&매도_호가_10, &매수_호가_10, &매도_호가_잔량_10, &매수_호가_잔량_10,
			&누적_거래량)
		lib.F테스트_에러없음(t, 에러)

		레코드값 := new(lib.NH호가_잔량)
		레코드값.M종목코드 = 코드
		레코드값.M시각 = 시각
		레코드값.M매도_호가_모음 = []int64{
			lib.F2정수64_단순형(매도_호가_1), lib.F2정수64_단순형(매도_호가_2), lib.F2정수64_단순형(매도_호가_3),
			lib.F2정수64_단순형(매도_호가_4), lib.F2정수64_단순형(매도_호가_5), lib.F2정수64_단순형(매도_호가_6),
			lib.F2정수64_단순형(매도_호가_7), lib.F2정수64_단순형(매도_호가_8), lib.F2정수64_단순형(매도_호가_9),
			lib.F2정수64_단순형(매도_호가_10)}
		레코드값.M매수_호가_모음 = []int64{
			lib.F2정수64_단순형(매수_호가_1), lib.F2정수64_단순형(매수_호가_2), lib.F2정수64_단순형(매수_호가_3),
			lib.F2정수64_단순형(매수_호가_4), lib.F2정수64_단순형(매수_호가_5), lib.F2정수64_단순형(매수_호가_6),
			lib.F2정수64_단순형(매수_호가_7), lib.F2정수64_단순형(매수_호가_8), lib.F2정수64_단순형(매수_호가_9),
			lib.F2정수64_단순형(매수_호가_10)}
		레코드값.M매도_잔량_모음 = []int64{
			매도_호가_잔량_1, 매도_호가_잔량_2, 매도_호가_잔량_3, 매도_호가_잔량_4, 매도_호가_잔량_5,
			매도_호가_잔량_6, 매도_호가_잔량_7, 매도_호가_잔량_8, 매도_호가_잔량_9, 매도_호가_잔량_10}
		레코드값.M매수_잔량_모음 = []int64{
			매수_호가_잔량_1, 매수_호가_잔량_2, 매수_호가_잔량_3, 매수_호가_잔량_4, 매수_호가_잔량_5,
			매수_호가_잔량_6, 매수_호가_잔량_7, 매수_호가_잔량_8, 매수_호가_잔량_9, 매수_호가_잔량_10}
		레코드값.M누적_거래량 = 누적_거래량
		

		일치하는_값_찾음 := false

		for _, 인터페이스_값 := range 값_모음 {
			값 := 인터페이스_값.(*lib.NH호가_잔량)

			if 값.M종목코드 == 레코드값.M종목코드 &&
				reflect.DeepEqual(값.M시각, 레코드값.M시각) &&
				값.M매도_호가_모음[0] == 레코드값.M매도_호가_모음[0] &&
				값.M매도_호가_모음[1] == 레코드값.M매도_호가_모음[1] &&
				값.M매도_호가_모음[2] == 레코드값.M매도_호가_모음[2] &&
				값.M매도_호가_모음[3] == 레코드값.M매도_호가_모음[3] &&
				값.M매도_호가_모음[4] == 레코드값.M매도_호가_모음[4] &&
				값.M매도_호가_모음[5] == 레코드값.M매도_호가_모음[5] &&
				값.M매도_호가_모음[6] == 레코드값.M매도_호가_모음[6] &&
				값.M매도_호가_모음[7] == 레코드값.M매도_호가_모음[7] &&
				값.M매도_호가_모음[8] == 레코드값.M매도_호가_모음[8] &&
				값.M매도_호가_모음[9] == 레코드값.M매도_호가_모음[9] &&
				값.M매수_호가_모음[0] == 레코드값.M매수_호가_모음[0] &&
				값.M매수_호가_모음[1] == 레코드값.M매수_호가_모음[1] &&
				값.M매수_호가_모음[2] == 레코드값.M매수_호가_모음[2] &&
				값.M매수_호가_모음[3] == 레코드값.M매수_호가_모음[3] &&
				값.M매수_호가_모음[4] == 레코드값.M매수_호가_모음[4] &&
				값.M매수_호가_모음[5] == 레코드값.M매수_호가_모음[5] &&
				값.M매수_호가_모음[6] == 레코드값.M매수_호가_모음[6] &&
				값.M매수_호가_모음[7] == 레코드값.M매수_호가_모음[7] &&
				값.M매수_호가_모음[8] == 레코드값.M매수_호가_모음[8] &&
				값.M매수_호가_모음[9] == 레코드값.M매수_호가_모음[9] &&
				값.M매도_잔량_모음[0] == 레코드값.M매도_잔량_모음[0] &&
				값.M매도_잔량_모음[1] == 레코드값.M매도_잔량_모음[1] &&
				값.M매도_잔량_모음[2] == 레코드값.M매도_잔량_모음[2] &&
				값.M매도_잔량_모음[3] == 레코드값.M매도_잔량_모음[3] &&
				값.M매도_잔량_모음[4] == 레코드값.M매도_잔량_모음[4] &&
				값.M매도_잔량_모음[5] == 레코드값.M매도_잔량_모음[5] &&
				값.M매도_잔량_모음[6] == 레코드값.M매도_잔량_모음[6] &&
				값.M매도_잔량_모음[7] == 레코드값.M매도_잔량_모음[7] &&
				값.M매도_잔량_모음[8] == 레코드값.M매도_잔량_모음[8] &&
				값.M매도_잔량_모음[9] == 레코드값.M매도_잔량_모음[9] &&
				값.M매수_잔량_모음[0] == 레코드값.M매수_잔량_모음[0] &&
				값.M매수_잔량_모음[1] == 레코드값.M매수_잔량_모음[1] &&
				값.M매수_잔량_모음[2] == 레코드값.M매수_잔량_모음[2] &&
				값.M매수_잔량_모음[3] == 레코드값.M매수_잔량_모음[3] &&
				값.M매수_잔량_모음[4] == 레코드값.M매수_잔량_모음[4] &&
				값.M매수_잔량_모음[5] == 레코드값.M매수_잔량_모음[5] &&
				값.M매수_잔량_모음[6] == 레코드값.M매수_잔량_모음[6] &&
				값.M매수_잔량_모음[7] == 레코드값.M매수_잔량_모음[7] &&
				값.M매수_잔량_모음[8] == 레코드값.M매수_잔량_모음[8] &&
				값.M매수_잔량_모음[9] == 레코드값.M매수_잔량_모음[9] &&
				값.M누적_거래량 == 누적_거래량 {
				일치하는_값_찾음 = true
				break
			}
		}

		lib.F테스트_참임(t, 일치하는_값_찾음)
	}
}

func Test실시간_데이터_저장_MySQL_시간외_호가잔량(t *testing.T) {
	const 테이블명 = "OffTimeOfferBid"

	값_생성함수 := func() interface{} {
		값 := new(lib.NH시간외_호가잔량)
		값.M종목코드 = 테스트용_종목코드
		값.M시각 = time.Now()
		값.M총_매도호가_잔량 = lib.F임의_범위_이내_정수64값(1, 1000000)
		값.M총_매수호가_잔량 = lib.F임의_범위_이내_정수64값(1, 10000000)

		return 값
	}

	값_모음 := f실시간_데이터_저장_MySQL_테스트_도우미(t, 테이블명, 테스트용_종목코드, 값_생성함수)

	버퍼 := new(bytes.Buffer)
	버퍼.WriteString("SELECT * FROM " + 테이블명 + " ")
	버퍼.WriteString("WHERE Code = ?")

	레코드_모음, 에러 := f질의_Rows(버퍼.String(), 테스트용_종목코드)
	lib.F테스트_에러없음(t, 에러)

	defer 레코드_모음.Close()

	for 레코드_모음.Next() {
		var id int64
		var 코드 string
		var 시각 time.Time
		var 매도호가잔량, 매수호가잔량 int64

		에러 = 레코드_모음.Scan(
			&id, &코드, &시각,
			&매도호가잔량, &매수호가잔량)

		lib.F테스트_에러없음(t, 에러)

		레코드값 := new(lib.NH시간외_호가잔량)
		레코드값.M종목코드 = 코드
		레코드값.M시각 = 시각
		레코드값.M총_매도호가_잔량 = 매도호가잔량
		레코드값.M총_매수호가_잔량 = 매수호가잔량

		일치하는_값_찾음 := false

		lib.F체크포인트(레코드값, "레코드")

		for i, 인터페이스_값 := range 값_모음 {
			값 := 인터페이스_값.(*lib.NH시간외_호가잔량)
			lib.F체크포인트(값, i)

			if 값.M종목코드 == 레코드값.M종목코드 &&
				값.M시각.Truncate(time.Millisecond) == 레코드값.M시각.Truncate(time.Millisecond) &&
				값.M총_매수호가_잔량 == 레코드값.M총_매수호가_잔량 &&
				값.M총_매도호가_잔량 == 레코드값.M총_매도호가_잔량 {
				일치하는_값_찾음 = true
				break
			}
		}

		lib.F테스트_참임(t, 일치하는_값_찾음)

		lib.F체크포인트(id)
	}
}

func Test실시간_데이터_저장_MySQL_예상_호가잔량(t *testing.T) {
	const 테이블명 = "EstimatedOfferBid"

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

	값_모음 := f실시간_데이터_저장_MySQL_테스트_도우미(t, 테이블명, 테스트용_종목코드, 값_생성함수)

	버퍼 := new(bytes.Buffer)
	버퍼.WriteString("SELECT * FROM " + 테이블명 + " ")
	버퍼.WriteString("WHERE Code = ?")

	레코드_모음, 에러 := f질의_Rows(버퍼.String(), 테스트용_종목코드)
	lib.F테스트_에러없음(t, 에러)

	defer 레코드_모음.Close()

	for 레코드_모음.Next() {
		var id int64
		var 코드 string
		var 시각 time.Time
		var 동시호가_구분, 예상_등락부호 uint8
		var 예상_체결가, 예상_등락폭, 매도_호가, 매수_호가 string
		var 예상_등락율 float64
		var 예상_체결수량, 매도_호가_잔량, 매수_호가_잔량 int64

		에러 = 레코드_모음.Scan(
			&id, &코드, &시각,
			&동시호가_구분, &예상_체결가, &예상_등락부호, &예상_등락폭,
			&예상_등락율, &예상_체결수량, &매도_호가, &매수_호가,
			&매도_호가_잔량, &매수_호가_잔량)

		lib.F테스트_에러없음(t, 에러)

		레코드값 := new(lib.NH예상_호가잔량)
		레코드값.M종목코드 = 코드
		레코드값.M시각 = 시각
		레코드값.M동시호가_구분 = 동시호가_구분
		레코드값.M예상_체결가 = lib.F2정수64_단순형(예상_체결가)
		레코드값.M예상_등락부호 = 예상_등락부호
		레코드값.M예상_등락폭 = lib.F2정수64_단순형(예상_등락폭)
		레코드값.M예상_등락율 = 예상_등락율
		레코드값.M예상_체결수량 = 예상_체결수량
		레코드값.M매도_호가 = lib.F2정수64_단순형(매도_호가)
		레코드값.M매수_호가 = lib.F2정수64_단순형(매수_호가)
		레코드값.M매도_호가잔량 = 매도_호가_잔량
		레코드값.M매수_호가잔량 = 매수_호가_잔량
		lib.F체크포인트(레코드값)

		일치하는_값_찾음 := false

		for i, 인터페이스_값 := range 값_모음 {
			값 := 인터페이스_값.(*lib.NH예상_호가잔량)
			lib.F체크포인트(값, i)

			if 값.M종목코드 == 레코드값.M종목코드 &&
				reflect.DeepEqual(값.M시각, 레코드값.M시각) &&
				값.M동시호가_구분 == 레코드값.M동시호가_구분 &&
				값.M예상_체결가 == 레코드값.M예상_체결가 &&
				값.M예상_등락부호 == 레코드값.M예상_등락부호 &&
				값.M예상_등락폭 == 레코드값.M예상_등락폭 &&
				lib.F비슷한_실수값(값.M예상_등락율, 레코드값.M예상_등락율) &&
				값.M예상_체결수량 == 레코드값.M예상_체결수량 &&
				값.M매도_호가 == 레코드값.M매도_호가 &&
				값.M매수_호가 == 레코드값.M매수_호가 &&
				값.M매도_호가잔량 == 레코드값.M매도_호가잔량 &&
				값.M매수_호가잔량 == 레코드값.M매수_호가잔량 {
				일치하는_값_찾음 = true
				break
			}
		}

		lib.F테스트_참임(t, 일치하는_값_찾음)
	}
}

func Test실시간_데이터_저장_MySQL_체결(t *testing.T) {
	const 테이블명 = "Deal"

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
		값.M거래대금_100만 = lib.F임의_범위_이내_정수64값(0, 1000000000)
		값.M시가 = lib.F임의_범위_이내_정수64값(1000, 10000000)
		값.M가중_평균_가격 = lib.F임의_범위_이내_정수64값(1000, 10000000)
		값.M시장구분 = lib.F임의_시장구분()

		return 값
	}

	값_모음 := f실시간_데이터_저장_MySQL_테스트_도우미(t, 테이블명, 테스트용_종목코드, 값_생성함수)

	버퍼 := new(bytes.Buffer)
	버퍼.WriteString("SELECT * FROM " + 테이블명 + " ")
	버퍼.WriteString("WHERE Code = ?")

	레코드_모음, 에러 := f질의_Rows(버퍼.String(), 테스트용_종목코드)
	lib.F테스트_에러없음(t, 에러)

	defer 레코드_모음.Close()

	for 레코드_모음.Next() {
		var id int64
		var 코드 string
		var 시각 time.Time
		var 등락부호 uint8
		var 등락폭 string
		var 등락율 float64
		var 현재가, 고가, 저가, 매도_호가, 매수_호가 string
		var 누적_거래량 int64
		var 전일대비_거래량_비율 float64
		var 변동_거래량 int64
		var 거래대금_백만, 시가, 가중_평균_가격 string
		var 시장구분 uint8

		에러 = 레코드_모음.Scan(
			&id, &코드, &시각,
			&등락부호, &등락폭, &현재가,
			&등락율, &고가, &저가, &매도_호가, &매수_호가,
			&누적_거래량, &전일대비_거래량_비율, &변동_거래량,
			&거래대금_백만, &시가, &가중_평균_가격, &시장구분)

		lib.F테스트_에러없음(t, 에러)

		레코드값 := new(lib.NH체결)
		레코드값.M종목코드 = 코드
		레코드값.M시각 = 시각
		레코드값.M등락부호 = 등락부호
		레코드값.M등락폭 = lib.F2정수64_단순형(등락폭)
		레코드값.M현재가 = lib.F2정수64_단순형(현재가)
		레코드값.M등락율 = 등락율
		레코드값.M고가 = lib.F2정수64_단순형(고가)
		레코드값.M저가 = lib.F2정수64_단순형(저가)
		레코드값.M매도_호가 = lib.F2정수64_단순형(매도_호가)
		레코드값.M매수_호가 = lib.F2정수64_단순형(매수_호가)
		레코드값.M누적_거래량  = 누적_거래량
		레코드값.M전일대비_거래량_비율 = 전일대비_거래량_비율
		레코드값.M변동_거래량 = 변동_거래량
		레코드값.M거래대금_100만 = lib.F2정수64_단순형(거래대금_백만)
		레코드값.M시가 = lib.F2정수64_단순형(시가)
		레코드값.M가중_평균_가격 = lib.F2정수64_단순형(가중_평균_가격)
		레코드값.M시장구분 = lib.T시장구분(시장구분)
		lib.F체크포인트(레코드값)

		일치하는_값_찾음 := false

		for i, 인터페이스_값 := range 값_모음 {
			값 := 인터페이스_값.(*lib.NH체결)
			lib.F체크포인트(값, i)

			if 값.M종목코드 == 레코드값.M종목코드 &&
				reflect.DeepEqual(값.M시각, 레코드값.M시각) &&
				값.M등락부호 == 레코드값.M등락부호 &&
				값.M등락폭 == 레코드값.M등락폭 &&
				값.M현재가 == 레코드값.M현재가 &&
				lib.F비슷한_실수값(값.M등락율, 레코드값.M등락율) &&
				값.M고가 == 레코드값.M고가 &&
				값.M저가 == 레코드값.M저가 &&
				값.M매도_호가 == 레코드값.M매도_호가 &&
				값.M매수_호가 == 레코드값.M매수_호가 &&
				값.M누적_거래량 == 레코드값.M누적_거래량 &&
				lib.F비슷한_실수값(값.M전일대비_거래량_비율, 레코드값.M전일대비_거래량_비율) &&
				값.M변동_거래량 == 레코드값.M변동_거래량 &&
				값.M거래대금_100만 == 레코드값.M거래대금_100만 &&
				값.M시가 == 레코드값.M시가 &&
				값.M가중_평균_가격 == 레코드값.M가중_평균_가격 &&
				값.M시장구분 == 레코드값.M시장구분 {
				일치하는_값_찾음 = true
				break
			}
		}

		lib.F테스트_참임(t, 일치하는_값_찾음)
	}
}

func Test실시간_데이터_저장_MySQL_ETF_NAV(t *testing.T) {
	const 테이블명 = "ETF_NAV"

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

	값_모음 := f실시간_데이터_저장_MySQL_테스트_도우미(t, 테이블명, 테스트용_종목코드, 값_생성함수)

	버퍼 := new(bytes.Buffer)
	버퍼.WriteString("SELECT * FROM " + 테이블명 + " ")
	버퍼.WriteString("WHERE Code = ?")

	레코드_모음, 에러 := f질의_Rows(버퍼.String(), 테스트용_종목코드)
	lib.F테스트_에러없음(t, 에러)

	defer 레코드_모음.Close()

	for 레코드_모음.Next() {
		var id int64
		var 코드 string
		var 시각 time.Time
		var 등락부호 uint8
		var 등락폭, 현재가_NAV, 시가_NAV, 고가_NAV, 저가_NAV  float64
		var 추적오차_부호 uint8
		var 추적오차    float64
		var 괴리율_부호  uint8
		var 괴리율     float64

		에러 = 레코드_모음.Scan(
			&id, &코드, &시각,
			&등락부호, &등락폭,
			&현재가_NAV, &시가_NAV, &고가_NAV, &저가_NAV,
			&추적오차_부호, &추적오차, &괴리율_부호, &괴리율)

		lib.F테스트_에러없음(t, 에러)

		레코드값 := new(lib.NH_ETF_NAV)
		레코드값.M종목코드 = 코드
		레코드값.M시각 = 시각
		레코드값.M등락부호 = 등락부호
		레코드값.M등락폭 = 등락폭
		레코드값.M현재가_NAV = 현재가_NAV
		레코드값.M시가_NAV = 시가_NAV
		레코드값.M고가_NAV = 고가_NAV
		레코드값.M저가_NAV = 저가_NAV
		레코드값.M추적오차_부호 = 추적오차_부호
		레코드값.M추적오차 = 추적오차
		레코드값.M괴리율_부호 = 괴리율_부호
		레코드값.M괴리율 = 괴리율

		일치하는_값_찾음 := false

		for _, 인터페이스_값 := range 값_모음 {
			값 := 인터페이스_값.(*lib.NH_ETF_NAV)

			if 값.M종목코드 == 레코드값.M종목코드 &&
				reflect.DeepEqual(값.M시각, 레코드값.M시각) &&
				값.M등락부호 == 레코드값.M등락부호 &&
				lib.F비슷한_실수값(값.M등락폭, 레코드값.M등락폭) &&
				lib.F비슷한_실수값(값.M현재가_NAV, 레코드값.M현재가_NAV) &&
				lib.F비슷한_실수값(값.M시가_NAV, 레코드값.M시가_NAV) &&
				lib.F비슷한_실수값(값.M고가_NAV, 레코드값.M고가_NAV) &&
				lib.F비슷한_실수값(값.M저가_NAV, 레코드값.M저가_NAV) &&
				값.M추적오차_부호 == 레코드값.M추적오차_부호 &&
				lib.F비슷한_실수값(값.M추적오차, 레코드값.M추적오차) &&
				값.M괴리율_부호 == 레코드값.M괴리율_부호 &&
				lib.F비슷한_실수값(값.M괴리율, 레코드값.M괴리율) {
				일치하는_값_찾음 = true
				break
			}
		}

		lib.F테스트_참임(t, 일치하는_값_찾음)
	}
}

func Test실시간_데이터_저장_MySQL_Sector_Index(t *testing.T) {
	const 테이블명 = "SectorIndex"

	값_생성함수 := func() interface{} {
		값 := new(lib.NH업종지수)
		값.M업종코드 = 테스트용_종목코드[:2]
		값.M시각 = time.Now()
		값.M현재값 = lib.F임의_범위_이내_실수64값(50, 10000)
		값.M등락부호 = uint8(lib.F임의_양의_정수8값())
		값.M등락폭 = lib.F임의_범위_이내_실수64값(50, 10000)
		값.M거래량 = lib.F임의_양의_정수64값()
		값.M거래대금 = lib.F임의_범위_이내_정수64값(0, 1000000000)
		값.M개장값 = lib.F임의_범위_이내_실수64값(50, 10000)
		값.M최고값 = lib.F임의_범위_이내_실수64값(50, 10000)
		값.M최고값_시각 = time.Now()
		값.M최저값 = lib.F임의_범위_이내_실수64값(50, 10000)
		값.M최저값_시각 = time.Now()
		값.M지수_등락율 = lib.F임의_범위_이내_실수64값(-100, 100)
		값.M거래비중 = lib.F임의_범위_이내_실수64값(0, 100)

		return 값
	}

	값_모음 := f실시간_데이터_저장_MySQL_테스트_도우미(t, 테이블명, 테스트용_종목코드[:2], 값_생성함수)

	버퍼 := new(bytes.Buffer)
	버퍼.WriteString("SELECT * FROM " + 테이블명 + " ")
	버퍼.WriteString("WHERE Code = ?")

	레코드_모음, 에러 := f질의_Rows(버퍼.String(), 테스트용_종목코드)
	lib.F테스트_에러없음(t, 에러)

	defer 레코드_모음.Close()

	for 레코드_모음.Next() {
		var id int64
		var 코드 string
		var 시각 time.Time
		var 현재값 float64
		var 등락부호   uint8
		var 등락폭    float64
		var 거래량    int64
		var 거래대금  int64
		var 개장값    float64
		var 최고값    float64
		var 최고값_시각 time.Time
		var 최저값    float64
		var 최저값_시각 time.Time
		var 지수_등락율 float64
		var 거래비중  float64

		에러 = 레코드_모음.Scan(
			&id, &코드, &시각,
			&현재값, &등락부호, &등락폭,
			&거래량, &거래대금, &개장값,
			&최고값, &최고값_시각, &최저값, &최저값_시각,
			&지수_등락율, &거래비중)

		lib.F테스트_에러없음(t, 에러)

		레코드값 := new(lib.NH업종지수)
		레코드값.M업종코드 = 코드
		레코드값.M시각 = 시각
		레코드값.M현재값 = 현재값
		레코드값.M등락부호 = 등락부호
		레코드값.M등락폭 = 등락폭
		레코드값.M거래량 = 거래량
		레코드값.M거래대금 = 거래대금
		레코드값.M개장값 = 개장값
		레코드값.M최고값 = 최고값
		레코드값.M최고값_시각 = 최고값_시각
		레코드값.M최저값 = 최저값
		레코드값.M최저값_시각 = 최저값_시각
		레코드값.M지수_등락율 = 지수_등락율
		레코드값.M거래비중 = 거래비중

		일치하는_값_찾음 := false

		for _, 인터페이스_값 := range 값_모음 {
			값 := 인터페이스_값.(*lib.NH업종지수)

			if 값.M업종코드 == 레코드값.M업종코드 &&
				reflect.DeepEqual(값.M시각, 레코드값.M시각) &&
				lib.F비슷한_실수값(값.M현재값, 레코드값.M현재값) &&
				값.M등락부호 == 레코드값.M등락부호 &&
				lib.F비슷한_실수값(값.M등락폭, 레코드값.M등락폭) &&
				값.M거래량 == 레코드값.M거래량 &&
				값.M거래대금 == 레코드값.M거래대금 &&
				값.M개장값 == 레코드값.M개장값 &&
				값.M최고값 == 레코드값.M최고값 &&
				reflect.DeepEqual(값.M최고값_시각, 레코드값.M최고값_시각) &&
				값.M최저값 == 레코드값.M최저값 &&
				reflect.DeepEqual(값.M최저값_시각, 레코드값.M최저값_시각) &&
				lib.F비슷한_실수값(값.M지수_등락율, 레코드값.M지수_등락율) &&
				lib.F비슷한_실수값(값.M거래비중, 레코드값.M거래비중) {
				일치하는_값_찾음 = true
				break
			}
		}

		lib.F테스트_참임(t, 일치하는_값_찾음)
	}
}
