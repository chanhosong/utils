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

package connector_relay

import (
	"github.com/ghts/ghts_connector/nh"
	"github.com/ghts/lib"
	"github.com/go-mangos/mangos"
	"time"
)

var 실시간_정보_중계_NH = lib.New안전한_bool(false)
var 구독내역_저장소_NH = new실시간_정보_구독_내역_저장소()
var 대기_중_데이터_저장소_NH = new대기_중_데이터_저장소()

func F질의_NH(질의값 interface{}, 추가_질의값 ...interface{}) (응답 lib.I소켓_메시지) {
	defer lib.F에러_패닉_처리(func(r interface{}) { 응답 = lib.New소켓_메시지_에러(r) })

	질의값_모음 := append([]interface{}{질의값}, 추가_질의값...)
	소켓_질의 := lib.New소켓_질의(lib.P주소_NH_TR, lib.CBOR, lib.P30초)
	소켓_질의.S질의(질의값_모음...)

	return 소켓_질의.G응답()
}

func F실시간_정보_구독_NH(SUB소켓 mangos.Socket, ch수신 chan lib.I소켓_메시지,
	RT코드 string, 종목_모음 []*lib.S종목) (에러 error) {
	defer lib.F에러_패닉_처리(&에러)

	F실시간_정보_중계_NH()

	// 타임아웃을 음수로 설정하면, non-blocking 동작을 한다.
	SUB소켓.SetOption(mangos.OptionRecvDeadline, -1 * time.Second)

	질의값 := nh.New실시간_정보_질의(RT코드, 종목_모음)
	응답 := F질의_NH(질의값)
	lib.F에러2패닉(응답.G에러())

	return 구독내역_저장소_NH.S추가(SUB소켓, ch수신)
}

func F실시간_정보_해지_NH(SUB소켓 mangos.Socket, ch수신 chan lib.I소켓_메시지,
	RT코드 string, 종목_모음 []*lib.S종목) (에러 error) {
	defer lib.F에러_패닉_처리(&에러)

	F실시간_정보_중계_NH()

	질의값 := nh.New실시간_정보_질의(RT코드, 종목_모음)
	응답 := F질의_NH(질의값)
	lib.F에러2패닉(응답.G에러())

	return 구독내역_저장소_NH.S삭제(SUB소켓, ch수신)
}

func F실시간_정보_중계_NH() {
	if 실시간_정보_중계_NH.G값() {
		return
	}

	go f실시간_정보_중계_NH()
}

func f실시간_정보_중계_NH() {
	if 에러 := 실시간_정보_중계_NH.S값(true); 에러 != nil {
		return
	}

	defer 실시간_정보_중계_NH.S값(false)

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
			바이트_모음, 에러 := 소켓.Recv()    // non-blocking 동작임.
			if len(바이트_모음) == 0 {
				continue
			} else if 에러 != nil {
				lib.F에러_출력(에러)
				continue
			}

			소켓_메시지 := lib.New소켓_메시지from바이트_모음(바이트_모음)
			if 소켓_메시지.G에러() != nil {
				lib.F에러_출력(소켓_메시지.G에러())
				continue
			}

			select {
			case ch수신 <- 소켓_메시지:   // 메시지 중계 성공
			default:                      // 메시지 중계 실패. 저장소에 보관하고 추후 재전송
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
