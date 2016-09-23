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
	NH "github.com/ghts/ghts_connector/nh"
	//Xing "github.com/ghts/ghts_connector/xing"
	//"github.com/boltdb/bolt"

	"time"
)

type S틱_데이터 struct {
	M종목코드     string
	M시각       time.Time
	M매수_호가_모음 []int64
	M매수_잔량_모음 []int64
	M매도_호가_모음 []int64
	M매도_잔량_모음 []int64
	M현재가      int64
	NAV float64
	M거래량      int64
}

func f틱_데이터_파일명(종목 *공용.S종목) string {
	return "tick_" + 종목.G코드() + "_" + time.Now().Format(공용.P일자_형식) + ".dat"
}

func f_ETF_틱_데이터_수집_NH(종목 *공용.S종목, 초기화_완료 chan bool) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case error:
				공용.F에러_출력(r.(error))
			default:
				공용.New에러("%v", r)
			}
		}
	}()

	// NH 루틴 시작
	if 에러 := NH.F초기화(); 에러 != nil {
		초기화_완료 <- false
		return
	}

	공용.F대기(공용.P500밀리초)

	// 소켓 준비
	TR소켓, 에러 := 공용.New소켓REQ(공용.P주소_NH_TR)
	공용.F에러_패닉(에러)
	defer TR소켓.Close()

	구독_소켓_CBOR, 에러 := 공용.New소켓SUB(공용.P주소_NH실시간_CBOR)
	공용.F에러_패닉(에러)
	defer 구독_소켓_CBOR.Close()

	초기화_완료 <- true

	// 기본 데이터 질의
	질의값 := NH.New조회_질의(NH.TR_ETF_현재가_조회, 종목)
	에러 = 공용.F송신zmq(TR소켓, 공용.P30초, 공용.CBOR, 공용.TR일반, 질의값)
	공용.F에러_패닉(에러)

	// 기본 데이터 수신
	응답 := 공용.F수신zmq(TR소켓, 공용.P30초)
	공용.F에러_패닉(응답.G에러())
	공용.F조건부_패닉(응답.G길이() != 1, "기본 데이터 응답 길이가 예상과 다릅니다. 예상 1, 실제  %v", 응답.G길이())

	조회값 := NH.NewNH_ETF_현재가_조회_응답()
	공용.F에러_패닉(응답.G값(0, 조회값))
	공용.F조건부_패닉(조회값.M기본_정보.M종목코드 != 종목.G코드(),
		"종목코드 불일치. %v %v", 조회값.M기본_정보.M종목코드 != 종목.G코드())

	초기_데이터 := new(S틱_데이터)
	초기_데이터.M종목코드 = 종목.G코드()
	초기_데이터.M시각 = 조회값.M기본_정보.M시각
	초기_데이터.M매수_호가_모음 = 조회값.M기본_정보.M매수_호가_모음
	초기_데이터.M매수_잔량_모음 = 조회값.M기본_정보.M매수_잔량_모음
	초기_데이터.M매도_호가_모음 = 조회값.M기본_정보.M매도_호가_모음
	초기_데이터.M매도_잔량_모음 = 조회값.M기본_정보.M매도_잔량_모음
	초기_데이터.M현재가 = 조회값.M기본_정보.M현재가
	초기_데이터.NAV = 조회값.ETF_정보.NAV
	초기_데이터.M거래량 = 조회값.M기본_정보.M거래량

	// 실시간 정보 질의 TR 송신
	질의값 = NH.New조회_질의(NH.RT코스피_체결, 종목)
	에러 = 공용.F송신zmq(TR소켓, 공용.P30초, 공용.CBOR, 공용.TR일반, 질의값)
	공용.F에러_패닉(에러)

	질의값 = NH.New조회_질의(NH.RT코스피_호가_잔량, 종목)
	에러 = 공용.F송신zmq(TR소켓, 공용.P30초, 공용.CBOR, 공용.TR일반, 질의값)
	공용.F에러_패닉(에러)

	질의값 = NH.New조회_질의(NH.RT코스피_ETF_NAV, 종목)
	에러 = 공용.F송신zmq(TR소켓, 공용.P30초, 공용.CBOR, 공용.TR일반, 질의값)
	공용.F에러_패닉(에러)

	// 실시간 데이터 수신
	for {
		if 공용.F공통_종료() {
			return
		}

		응답 = 공용.F수신zmq(구독_소켓_CBOR, 공용.P1초)
		if 응답.G에러() != nil {
			공용.F메모(응답.G에러().Error())
			continue
		}

		// TODO
		for i := 0; i < 응답.G길이(); i++ {
			공용.F메모(응답.G자료형_문자열(i))
			//switch 응답.G자료형_문자열() {
			//default:
			//	공용.F메모(응답.G자료형_문자열())
			//}
		}
	}

	// 데이터 저장 (bolt)
	// TODO
}