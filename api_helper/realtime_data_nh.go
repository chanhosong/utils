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

package api_helper

import (
	"github.com/ghts/lib"
	"time"
)

var 실시간_정보_중계_NH = lib.New안전한_bool(false)
var 구독내역_저장소_NH = new실시간_정보_구독_내역_저장소()
var 대기_중_데이터_저장소_NH = new대기_중_데이터_저장소()

func fNH_실시간_정보_중계_초기화() {
	if !실시간_정보_중계_NH.G값() {
		ch초기화 := make(chan lib.T신호)
		go f실시간_정보_중계_NH(ch초기화)
		<-ch초기화
	}
}

func f실시간_정보_중계_NH(ch초기화 chan lib.T신호) {
	if 에러 := 실시간_정보_중계_NH.S값(true); 에러 != nil {
		ch초기화 <- lib.P신호_초기화
		return
	}

	defer 실시간_정보_중계_NH.S값(false)
	ch초기화 <- lib.P신호_초기화

	for {
		for _, 소켓 := range 구독내역_저장소_NH.G소켓_모음() {
			lib.F메모("소켓과 수신 채널이 유효한 지 검사해야 함.")

			if 소켓 == nil {
				continue
			}

			ch수신, 에러 := 구독내역_저장소_NH.G중계_채널(소켓)
			if ch수신 == nil || 에러 != nil {
				continue
			}

			// 해당 소켓에 대기 중인 모든 메시지를 중계한 후 다음 소켓으로 넘어간다.
			// 타임아웃을 음수로 설정해서 non-blocking으로 동작함.
			바이트_모음, 에러 := 소켓.Recv()
			if 에러 != nil {
				lib.F에러_출력(에러)
				continue
			} else if len(바이트_모음) == 0 {
				continue
			}

			소켓_메시지 := lib.New소켓_메시지from바이트_모음(바이트_모음)
			if 소켓_메시지.G에러() != nil {
				lib.F에러_출력(소켓_메시지.G에러())
				continue
			}

			select {
			case ch수신 <- 소켓_메시지: // 메시지 중계 성공
			default: // 메시지 중계 실패. 저장소에 보관하고 추후 재전송
				대기_중_데이터_저장소_NH.S추가(소켓_메시지, ch수신)
			}
		}

		대기_중_데이터_저장소_NH.S재전송()

		// 종료 조건 확인
		select {
		case <-lib.F공통_종료_채널():
			return
		}
	}
}

var 실시간_데이터_수집_NH_ETF = lib.New안전한_bool(false)
var ch실시간_데이터_수집 = make(chan lib.I소켓_메시지, 100000)

var nh호가_잔량 = lib.F자료형_문자열(lib.NH호가_잔량{})
var nh시간외_호가잔량 = lib.F자료형_문자열(lib.NH시간외_호가잔량{})
var nh예상_호가잔량 = lib.F자료형_문자열(lib.NH예상_호가잔량{})
var nh체결 = lib.F자료형_문자열(lib.NH체결{})
var nh_ETF_NAV = lib.F자료형_문자열(lib.NH_ETF_NAV{})
var nh업종지수 = lib.F자료형_문자열(lib.NH업종지수{})

var 버킷_호가_잔량 = []byte("RtBidOffer")
var 버킷_시간외_호가_잔량 = []byte("RtOfftimeBidOffer")
var 버킷_예상_호가_잔량 = []byte("RtProjectedBidOffer")
var 버킷_체결 = []byte("RtTrade")
var 버킷_ETF_NAV = []byte("RtEtfNav")
var 버킷_업종지수 = []byte("RtSectorIndex")

func fNH_실시간_데이터_수집_초기화() {
	if !실시간_데이터_수집_NH_ETF.G값() {
		ch초기화 := make(chan lib.T신호)
		go f실시간_데이터_저장(ch초기화)
		<-ch초기화
	}
}

func f실시간_데이터_저장(ch초기화 chan lib.T신호) {
	if 에러 := 실시간_데이터_수집_NH_ETF.S값(true); 에러 != nil {
		ch초기화 <- lib.P신호_초기화
		return
	}

	defer 실시간_데이터_수집_NH_ETF.S값(false)

	데이터베이스, 에러 := lib.NewBoltDB(fNH_실시간_데이터_파일명())
	lib.F에러2패닉(에러)

	var 수신_메시지 lib.I소켓_메시지
	ch종료 := lib.F공통_종료_채널()
	ch초기화 <- lib.P신호_초기화

	// 실시간 데이터 수신
	for {
		select {
		case 수신_메시지 = <-ch실시간_데이터_수집:
			if 수신_메시지.G에러() != nil {
				lib.F에러_출력(수신_메시지.G에러())
				continue
			}
		case <-ch종료:
			return
		}

		// DB에 수신값 저장
		if 에러 = f실시간_데이터_저장_도우미(수신_메시지, 데이터베이스); 에러 != nil {
			lib.F에러_출력(에러)
		}
	}
}

func f실시간_데이터_저장_도우미(수신_메시지 lib.I소켓_메시지, 데이터베이스 lib.I데이터베이스) (에러 error) {
	defer lib.F에러패닉_처리(lib.S에러패닉_처리{M에러: &에러})

	lib.F조건부_패닉(수신_메시지.G길이() != 1, "예상하지 못한 메시지 길이. %v", 수신_메시지.G길이())

	질의 := new(lib.S데이터베이스_질의)

	switch 수신_메시지.G자료형_문자열(0) {
	case nh호가_잔량:
		s := new(lib.NH호가_잔량)
		lib.F에러2패닉(수신_메시지.G값(0, s))

		시각, 에러 := s.M시각.MarshalBinary()
		lib.F에러2패닉(에러)

		질의.M버킷ID = 버킷_호가_잔량
		질의.M키 = append([]byte(s.M종목코드), 시각...)
	case nh시간외_호가잔량:
		s := new(lib.NH시간외_호가잔량)
		lib.F에러2패닉(수신_메시지.G값(0, s))

		시각, 에러 := s.M시각.MarshalBinary()
		lib.F에러2패닉(에러)

		질의.M버킷ID = 버킷_시간외_호가_잔량
		질의.M키 = append([]byte(s.M종목코드), 시각...)
	case nh예상_호가잔량:
		s := new(lib.NH예상_호가잔량)
		lib.F에러2패닉(수신_메시지.G값(0, s))

		시각, 에러 := s.M시각.MarshalBinary()
		lib.F에러2패닉(에러)

		질의.M버킷ID = 버킷_예상_호가_잔량
		질의.M키 = append([]byte(s.M종목코드), 시각...)
	case nh체결:
		s := new(lib.NH체결)
		lib.F에러2패닉(수신_메시지.G값(0, s))

		시각, 에러 := s.M시각.MarshalBinary()
		lib.F에러2패닉(에러)

		질의.M버킷ID = 버킷_체결
		질의.M키 = append([]byte(s.M종목코드), 시각...)
	case nh_ETF_NAV:
		s := new(lib.NH_ETF_NAV)
		lib.F에러2패닉(수신_메시지.G값(0, s))

		시각, 에러 := s.M시각.MarshalBinary()
		lib.F에러2패닉(에러)

		질의.M버킷ID = 버킷_ETF_NAV
		질의.M키 = append([]byte(s.M종목코드), 시각...)
	case nh업종지수:
		s := new(lib.NH업종지수)
		lib.F에러2패닉(수신_메시지.G값(0, s))

		시각, 에러 := s.M시각.MarshalBinary()
		lib.F에러2패닉(에러)

		질의.M버킷ID = 버킷_업종지수
		질의.M키 = append([]byte(s.M업종코드), 시각...)
	default:
		lib.F패닉("예상하지 못한 자료형. %v", 수신_메시지.G자료형_문자열(0))
	}

	질의.M값, 에러 = 수신_메시지.G바이트_모음(0)
	lib.F에러2패닉(에러)

	return 데이터베이스.S업데이트(질의)
}

func fNH_실시간_데이터_파일명() string {
	return "realtime_data_NH_" + time.Now().Format(lib.P일자_형식) + ".dat"
}
