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
	nh "github.com/ghts/api_helper_nh"
	_ "github.com/go-sql-driver/mysql"
	"bytes"
	"database/sql"
	"time"
)

var 실시간_데이터_수집_MySQL_고루틴_실행_중 = lib.New안전한_bool(false)

func f실시간_데이터_수집_NH_ETF_MySQL(종목코드_모음 []string) (db *sql.DB, 에러 error) {
	defer lib.F에러패닉_처리(lib.S에러패닉_처리{
		M에러: &에러,
		M함수with패닉내역: func(r interface{}) {
			lib.New에러with출력(r)
			db = nil
		}})

	lib.F체크포인트()

	ch수신 := make(chan lib.I소켓_메시지, 10000)
	ch초기화 := make(chan lib.T신호)

	lib.F체크포인트("1")

	nh.F실시간_데이터_구독_NH_ETF(ch수신, 종목코드_모음)

	lib.F체크포인트("2")

	defer nh.F실시간_데이터_해지_NH_ETF(종목코드_모음)

	lib.F체크포인트("3")

	nh.Go루틴_실시간_정보_중계_MySQL(ch초기화)
	신호 := <-ch초기화
	lib.F조건부_패닉(신호 != lib.P신호_초기화, "예상하지 못한 신호. %v", 신호)

	lib.F체크포인트()

	go go루틴_실시간_데이터_수집_MySQL(ch초기화, ch수신)
	신호 = <-ch초기화
	lib.F조건부_패닉(신호 != lib.P신호_초기화, "예상하지 못한 신호. %v", 신호)

	lib.F체크포인트()

	return db, nil
}

func go루틴_실시간_데이터_수집_MySQL(ch초기화 chan lib.T신호, ch수신 chan lib.I소켓_메시지) {
	var 에러 error
	lib.F에러패닉_처리(lib.S에러패닉_처리{M에러: &에러})

	if 에러 = 실시간_데이터_수집_MySQL_고루틴_실행_중.S값(true); 에러 != nil {
		ch초기화 <- lib.P신호_초기화
		lib.F패닉(에러)
	}

	defer 실시간_데이터_수집_MySQL_고루틴_실행_중.S값(false)

	var 수신_메시지 lib.I소켓_메시지
	ch종료 := lib.F공통_종료_채널()
	ch초기화 <- lib.P신호_초기화
	ch매1초 := time.Tick(lib.P1초)
	ch대기열 := make(chan lib.I소켓_메시지)

	// 실시간 데이터 수신
	for {
		select {
		case 수신_메시지 = <-ch수신:
			if 수신_메시지.G에러() != nil {
				lib.F에러_출력(수신_메시지.G에러())
				continue
			}

			if ch대기열 <- 수신_메시지; len(ch대기열) >= 1000 {
				lib.F에러2패닉(fNH_실시간_데이터_저장_MySQL(ch대기열))
			}
		case <-ch매1초:
			lib.F에러2패닉(fNH_실시간_데이터_저장_MySQL(ch대기열))
		case <-ch종료:
			return
		}
	}
}

func fNH_실시간_데이터_저장_MySQL(ch대기열 chan lib.I소켓_메시지) (에러 error) {
	var tx *sql.Tx = nil
	var 롤백_해야함 = false

	defer lib.F에러패닉_처리(lib.S에러패닉_처리{
		M에러: &에러,
		M함수: func() {
			if tx != nil && 롤백_해야함 {
				lib.F에러2패닉(tx.Rollback())
			}}})

	db, 에러 := fMySQL_DB()
	lib.F에러2패닉(에러)

	tx, 에러 = db.Begin()
	lib.F에러2패닉(에러)

	버퍼 := new(bytes.Buffer)
	버퍼.WriteString("INSERT INTO OfferBid (")
	버퍼.WriteString( "Code, Time,")
	버퍼.WriteString( "OfferPrice1, BidPrice1, OfferVolume1, BidVolume1,")
	버퍼.WriteString( "OfferPrice2, BidPrice2, OfferVolume2, BidVolume2,")
	버퍼.WriteString( "OfferPrice3, BidPrice3, OfferVolume3, BidVolume3,")
	버퍼.WriteString( "OfferPrice4, BidPrice4, OfferVolume4, BidVolume4,")
	버퍼.WriteString( "OfferPrice5, BidPrice5, OfferVolume5, BidVolume5,")
	버퍼.WriteString( "OfferPrice6, BidPrice6, OfferVolume6, BidVolume6,")
	버퍼.WriteString( "OfferPrice7, BidPrice7, OfferVolume7, BidVolume7,")
	버퍼.WriteString( "OfferPrice8, BidPrice8, OfferVolume8, BidVolume8,")
	버퍼.WriteString( "OfferPrice9, BidPrice9, OfferVolume9, BidVolume9,")
	버퍼.WriteString( "OfferPrice10, BidPrice10, OfferVolume10, BidVolume10,")
	버퍼.WriteString( "Volume")
	버퍼.WriteString(") VALUES (")
	버퍼.WriteString("?, ?, ?, ?, ?, ?, ?, ?, ?, ?,")
	버퍼.WriteString("?, ?, ?, ?, ?, ?, ?, ?, ?, ?,")
	버퍼.WriteString("?, ?, ?, ?, ?, ?, ?, ?, ?, ?,")
	버퍼.WriteString("?, ?, ?, ?, ?, ?, ?, ?, ?, ?,")
	버퍼.WriteString("?, ?, ?")
	버퍼.WriteString(")")
	stmtNH호가잔량, 에러 := tx.Prepare(버퍼.String())
	lib.F에러2패닉(에러)

	버퍼 = new(bytes.Buffer)
	버퍼.WriteString("INSERT INTO OffTimeOfferBid (")
	버퍼.WriteString( "Code,")
	버퍼.WriteString( "Time,")
	버퍼.WriteString( "OfferVolume,")
	버퍼.WriteString( "BidVolume")
	버퍼.WriteString(") VALUES (")
	버퍼.WriteString("?, ?, ?, ?")
	버퍼.WriteString(")")
	stmtNH시간외_호가잔량, 에러 := tx.Prepare(버퍼.String())
	lib.F에러2패닉(에러)

	버퍼 = new(bytes.Buffer)
	버퍼.WriteString("INSERT INTO EstimatedOfferBid (")
	버퍼.WriteString( "Code,")
	버퍼.WriteString( "Time,")
	버퍼.WriteString( "SyncOfferBid,")
	버퍼.WriteString( "EstmPrice,")
	버퍼.WriteString( "EstmDiffSign,")
	버퍼.WriteString( "EstmDiff,")
	버퍼.WriteString( "EstmDiffRate,")
	버퍼.WriteString( "EstmVolume,")
	버퍼.WriteString( "OfferPrice,")
	버퍼.WriteString( "BidPrice,")
	버퍼.WriteString( "OfferVolume,")
	버퍼.WriteString( "BidVolume")
	버퍼.WriteString(") VALUES (")
	버퍼.WriteString("?, ?, ?, ?, ?,")
	버퍼.WriteString("?, ?, ?, ?, ?,")
	버퍼.WriteString("?, ?")
	버퍼.WriteString(")")
	stmtNH예상_호가잔량, 에러 := tx.Prepare(버퍼.String())
	lib.F에러2패닉(에러)

	버퍼 = new(bytes.Buffer)
	버퍼.WriteString("INSERT INTO Deal (")
	버퍼.WriteString( "Code,")
	버퍼.WriteString( "Time,")
	버퍼.WriteString( "DiffSign,")
	버퍼.WriteString( "Diff,")
	버퍼.WriteString( "MarketPrice,")
	버퍼.WriteString( "DiffRate,")
	버퍼.WriteString( "High,")
	버퍼.WriteString( "Low,")
	버퍼.WriteString( "OfferPrice,")
	버퍼.WriteString( "BidPrice,")
	버퍼.WriteString( "Volume,")
	버퍼.WriteString( "VsPrevVolRate,")
	버퍼.WriteString( "DiffVolume,")
	버퍼.WriteString( "TrAmount,")
	버퍼.WriteString( "Open,")
	버퍼.WriteString( "WeightAvgPrice,")
	버퍼.WriteString( "Market")
	버퍼.WriteString(") VALUES (")
	버퍼.WriteString("?, ?, ?, ?, ?,")
	버퍼.WriteString("?, ?, ?, ?, ?,")
	버퍼.WriteString("?, ?, ?, ?, ?,")
	버퍼.WriteString("?, ?")
	버퍼.WriteString(")")
	stmtNH체결, 에러 := tx.Prepare(버퍼.String())
	lib.F에러2패닉(에러)

	버퍼 = new(bytes.Buffer)
	버퍼.WriteString("INSERT INTO ETF_NAV (")
	버퍼.WriteString( "Code,")
	버퍼.WriteString( "Time,")
	버퍼.WriteString( "DiffSign,")
	버퍼.WriteString( "Diff,")
	버퍼.WriteString( "Current,")
	버퍼.WriteString( "Open,")
	버퍼.WriteString( "High,")
	버퍼.WriteString( "Low,")
	버퍼.WriteString( "TrackErrSign,")
	버퍼.WriteString( "TrackingError,")
	버퍼.WriteString( "DivergeSign,")
	버퍼.WriteString( "DivergeRate")
	버퍼.WriteString(") VALUES (")
	버퍼.WriteString("?, ?, ?, ?, ?,")
	버퍼.WriteString("?, ?, ?, ?, ?,")
	버퍼.WriteString("?, ?")
	버퍼.WriteString(")")
	stmtNH_ETF_NAV, 에러 := tx.Prepare(버퍼.String())
	lib.F에러2패닉(에러)

	버퍼 = new(bytes.Buffer)
	버퍼.WriteString("INSERT INTO SectorIndex (")
	버퍼.WriteString( "Code,")
	버퍼.WriteString( "Time,")
	버퍼.WriteString( "IndexValue,")
	버퍼.WriteString( "DiffSign,")
	버퍼.WriteString( "Diff,")
	버퍼.WriteString( "Volume,")
	버퍼.WriteString( "TrAmount,")
	버퍼.WriteString( "Open,")
	버퍼.WriteString( "High,")
	버퍼.WriteString( "HighTime,")
	버퍼.WriteString( "Low,")
	버퍼.WriteString( "LowTime,")
	버퍼.WriteString( "DiffRate,")
	버퍼.WriteString( "TrVolRate")
	버퍼.WriteString(") VALUES (")
	버퍼.WriteString("?, ?, ?, ?, ?,")
	버퍼.WriteString("?, ?, ?, ?, ?,")
	버퍼.WriteString("?, ?, ?, ?")
	버퍼.WriteString(")")
	stmtNH업종지수, 에러 := tx.Prepare(버퍼.String())
	lib.F에러2패닉(에러)

	NH호가_잔량 := lib.F자료형_문자열(lib.NH호가_잔량{})
	NH시간외_호가잔량 := lib.F자료형_문자열(lib.NH시간외_호가잔량{})
	NH예상_호가잔량 := lib.F자료형_문자열(lib.NH예상_호가잔량{})
	NH체결 := lib.F자료형_문자열(lib.NH체결{})
	NH_ETF_NAV := lib.F자료형_문자열(lib.NH_ETF_NAV{})
	NH업종지수 := lib.F자료형_문자열(lib.NH업종지수{})

	롤백_해야함 = true
	길이 := len(ch대기열)

	for i:=0 ; i<길이 ; i++ {
		수신_메시지 := <-ch대기열

		lib.F조건부_패닉(수신_메시지 == nil, "nil 메시지")
		lib.F조건부_패닉(수신_메시지.G길이() != 1, "예상하지 못한 메시지 길이. %v", 수신_메시지.G길이())

		switch 수신_메시지.G자료형_문자열(0) {
		case NH호가_잔량:
			s := new(lib.NH호가_잔량)
			lib.F에러2패닉(수신_메시지.G값(0, s))

			lib.F조건부_패닉(len(s.M매도_잔량_모음) < 10, "매도 잔량 모음 데이터 부족", len(s.M매도_잔량_모음))
			lib.F조건부_패닉(len(s.M매도_호가_모음) < 10, "매도 호가 모음 데이터 부족", len(s.M매도_호가_모음))
			lib.F조건부_패닉(len(s.M매수_잔량_모음) < 10, "매수 잔량 모음 데이터 부족", len(s.M매수_잔량_모음))
			lib.F조건부_패닉(len(s.M매수_호가_모음) < 10, "매수 호가 모음 데이터 부족", len(s.M매수_호가_모음))

			_, 에러 = stmtNH호가잔량.Exec(
				s.M종목코드,
				s.M시각,
				s.M매도_호가_모음[0], s.M매수_호가_모음[0], s.M매도_잔량_모음[0], s.M매수_잔량_모음[0],
				s.M매도_호가_모음[1], s.M매수_호가_모음[1], s.M매도_잔량_모음[1], s.M매수_잔량_모음[1],
				s.M매도_호가_모음[2], s.M매수_호가_모음[2], s.M매도_잔량_모음[2], s.M매수_잔량_모음[2],
				s.M매도_호가_모음[3], s.M매수_호가_모음[3], s.M매도_잔량_모음[3], s.M매수_잔량_모음[3],
				s.M매도_호가_모음[4], s.M매수_호가_모음[4], s.M매도_잔량_모음[4], s.M매수_잔량_모음[4],
				s.M매도_호가_모음[5], s.M매수_호가_모음[5], s.M매도_잔량_모음[5], s.M매수_잔량_모음[5],
				s.M매도_호가_모음[6], s.M매수_호가_모음[6], s.M매도_잔량_모음[6], s.M매수_잔량_모음[6],
				s.M매도_호가_모음[7], s.M매수_호가_모음[7], s.M매도_잔량_모음[7], s.M매수_잔량_모음[7],
				s.M매도_호가_모음[8], s.M매수_호가_모음[8], s.M매도_잔량_모음[8], s.M매수_잔량_모음[8],
				s.M매도_호가_모음[9], s.M매수_호가_모음[9], s.M매도_잔량_모음[9], s.M매수_잔량_모음[9],
				s.M누적_거래량)

			lib.F에러2패닉(에러)
		case NH시간외_호가잔량:
			s := new(lib.NH시간외_호가잔량)
			lib.F에러2패닉(수신_메시지.G값(0, s))

			_, 에러 = stmtNH시간외_호가잔량.Exec(
				s.M종목코드,
				s.M시각,
				s.M총_매도호가_잔량,
				s.M총_매수호가_잔량)

			lib.F에러2패닉(에러)
		case NH예상_호가잔량:
			s := new(lib.NH예상_호가잔량)
			lib.F에러2패닉(수신_메시지.G값(0, s))

			_, 에러 = stmtNH예상_호가잔량.Exec(
				s.M종목코드,
				s.M시각,
				s.M동시호가_구분,
				s.M예상_체결가,
				s.M예상_등락부호,
				s.M예상_등락폭,
				s.M예상_등락율,
				s.M예상_체결수량,
				s.M매도_호가,
				s.M매수_호가,
				s.M매도_호가잔량,
				s.M매수_호가잔량)

			lib.F에러2패닉(에러)
		case NH체결:
			s := new(lib.NH체결)
			lib.F에러2패닉(수신_메시지.G값(0, s))

			_, 에러 = stmtNH체결.Exec(
				s.M종목코드,
				s.M시각,
				s.M등락부호,
				s.M등락폭,
				s.M현재가,
				s.M등락율,
				s.M고가,
				s.M저가,
				s.M매도_호가,
				s.M매수_호가,
				s.M누적_거래량,
				s.M전일대비_거래량_비율,
				s.M변동_거래량,
				s.M거래_대금_백만,
				s.M시가,
				s.M가중_평균_가격,
				s.M시장구분)

			lib.F에러2패닉(에러)
		case NH_ETF_NAV:
			s := new(lib.NH_ETF_NAV)
			lib.F에러2패닉(수신_메시지.G값(0, s))

			_, 에러 = stmtNH_ETF_NAV.Exec(
				s.M종목코드,
				s.M시각,
				s.M등락부호,
				s.M등락폭,
				s.M현재가_NAV,
				s.M시가_NAV,
				s.M고가_NAV,
				s.M저가_NAV,
				s.M추적오차_부호,
				s.M추적오차,
				s.M괴리율_부호,
				s.M괴리율)

			lib.F에러2패닉(에러)
		case NH업종지수:
			s := new(lib.NH업종지수)
			lib.F에러2패닉(수신_메시지.G값(0, s))

			_, 에러 = stmtNH업종지수.Exec(
				s.M업종코드,
				s.M시각,
				s.M현재값,
				s.M등락부호,
				s.M등락폭,
				s.M거래량,
				s.M거래대금,
				s.M개장값,
				s.M최고값,
				s.M최고값_시각,
				s.M최저값,
				s.M최저값_시각,
				s.M지수_등락율,
				s.M거래비중)

			lib.F에러2패닉(에러)
		default:
			lib.F패닉("예상하지 못한 자료형. %v", 수신_메시지.G자료형_문자열(0))
		}
	}

	lib.F에러2패닉(tx.Commit())

	return nil
}