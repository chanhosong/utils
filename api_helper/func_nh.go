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
	"github.com/go-mangos/mangos"
)

// 접속 되었는 지 확인.
func F접속됨_NH() (참거짓 bool) {
	defer lib.F에러패닉_처리(lib.S에러패닉_처리{M함수: func() { 참거짓 = false }})

	응답 := F질의_NH(lib.TR접속됨)
	lib.F에러2패닉(응답.G에러())
	lib.F에러2패닉(응답.G값(0, &참거짓))

	return 참거짓
}

func F질의_NH(TR구분 lib.TR구분, 질의값_모음 ...interface{}) (응답 lib.I소켓_메시지) {
	defer lib.F에러패닉_처리(lib.S에러패닉_처리{M함수with패닉내역: func(r interface{}) { 응답 = lib.New소켓_메시지_에러(r) }})

	소켓_질의 := lib.New소켓_질의(lib.P주소_NH_TR, lib.CBOR, lib.P30초)
	질의값_모음 = append([]interface{}{TR구분}, 질의값_모음...)

	return 소켓_질의.S질의(질의값_모음...).G응답()
}

func F실시간_정보_구독_NH(ch수신 chan lib.I소켓_메시지, RT코드 string, 종목코드_모음 []string) (에러 error) {
	defer lib.F에러패닉_처리(lib.S에러패닉_처리{M에러: &에러})

	fNH_실시간_정보_중계_초기화()

	질의값 := lib.NewNH실시간_정보_질의값(RT코드, 종목코드_모음)
	응답 := F질의_NH(lib.TR실시간_정보_구독, 질의값)
	lib.F에러2패닉(응답.G에러())

	구독_소켓 := 구독내역_저장소_NH.G소켓(ch수신)
	if 구독_소켓 != nil {
		return nil
	}

	구독_소켓, 에러 = lib.New소켓SUB(lib.P주소_NH실시간_CBOR)
	lib.F에러2패닉(에러)

	// 음수 타임아웃값은 non-blocking을 의미
	구독_소켓.SetOption(mangos.OptionRecvDeadline, lib.P마이너스1초)

	return 구독내역_저장소_NH.S추가(ch수신, 구독_소켓)
}

func F실시간_정보_해지_NH(ch수신 chan lib.I소켓_메시지, RT코드 string, 종목코드_모음 []string) (에러 error) {
	defer lib.F에러패닉_처리(lib.S에러패닉_처리{M에러: &에러})

	질의값 := lib.NewNH실시간_정보_질의값(RT코드, 종목코드_모음)
	응답 := F질의_NH(lib.TR실시간_정보_해지, 질의값)
	//구독내역_저장소_NH.S삭제(ch수신) // 등록된 소켓을 삭제하면 안 될 것 같다.
	return 응답.G에러()
}

func F실시간_데이터_수집_NH_ETF(종목코드_모음 []string) {
	var 에러 error

	defer lib.F에러패닉_처리(lib.S에러패닉_처리{
		M에러: &에러,
		M함수with패닉내역: func(r interface{}) {
			lib.New에러with출력(r)
		}})

	fNH_실시간_데이터_수집_초기화()

	lib.F대기(lib.P500밀리초)
	lib.F에러2패닉(F실시간_정보_구독_NH(ch실시간_데이터_수집, lib.NH_RT코스피_호가_잔량, 종목코드_모음))
	lib.F에러2패닉(F실시간_정보_구독_NH(ch실시간_데이터_수집, lib.NH_RT코스피_체결, 종목코드_모음))
	lib.F에러2패닉(F실시간_정보_구독_NH(ch실시간_데이터_수집, lib.NH_RT코스피_ETF_NAV, 종목코드_모음))
}
