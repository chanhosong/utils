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
	"testing"
	"strings"
	"time"
	"bytes"
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

func Test실시간_데이터_저장_MySQL_호가잔량(t *testing.T) {
	db, 에러 := fMySQL_DB()
	lib.F에러2패닉(에러)

	버퍼 := new(bytes.Buffer)
	버퍼.WriteString("SELECT COUNT(*) FROM OfferBid ")
	버퍼.WriteString("WHERE Code = ? ")

	var 원래_수량 int
	에러 = db.QueryRow(버퍼.String(), "TEST").Scan(&원래_수량)
	lib.F테스트_에러없음(t, 에러)

	호가잔량1 := new(lib.NH호가_잔량)
	호가잔량1.M종목코드 = "TEST"
	호가잔량1.M시각 = time.Now()
	호가잔량1.M매도_호가_모음 = []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	호가잔량1.M매도_잔량_모음 = []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	호가잔량1.M매수_호가_모음 = []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	호가잔량1.M매수_잔량_모음 = []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	호가잔량1.M누적_거래량 = 0

	소켓_메시지1, 에러 := lib.New소켓_메시지(lib.CBOR, 호가잔량1)
	lib.F테스트_에러없음(t, 에러)

	호가잔량2 := new(lib.NH호가_잔량)
	호가잔량2.M종목코드 = "TEST"
	호가잔량2.M시각 = time.Now()
	호가잔량2.M매도_호가_모음 = []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	호가잔량2.M매도_잔량_모음 = []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	호가잔량2.M매수_호가_모음 = []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	호가잔량2.M매수_잔량_모음 = []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	호가잔량2.M누적_거래량 = 0

	소켓_메시지2, 에러 := lib.New소켓_메시지(lib.CBOR, 호가잔량2)
	lib.F테스트_에러없음(t, 에러)

	ch대기열 := make(chan lib.I소켓_메시지, 10)
	ch대기열 <- 소켓_메시지1
	ch대기열 <- 소켓_메시지2

	에러 = fNH_실시간_데이터_저장_MySQL(ch대기열)
	lib.F테스트_에러없음(t, 에러)

	var 추가_후_수량 int
	에러 = db.QueryRow(버퍼.String(), "TEST").Scan(&추가_후_수량)
	lib.F테스트_에러없음(t, 에러)

	lib.F테스트_같음(t, 추가_후_수량, 원래_수량 + 2)
}

func TestF실시간_데이터_수집_MySQL(t *testing.T) {
	종목코드_모음 := lib.F2종목코드_모음(lib.F샘플_종목_모음_ETF())

	db, 에러 := f실시간_데이터_수집_NH_ETF_MySQL(종목코드_모음)
	defer db.Close()

	lib.F테스트_에러없음(t, 에러)
}