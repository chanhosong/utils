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
	"testing"
)

func TestF질의TR_NH(t *testing.T) {
	TR구분 := lib.TR일반
	질의값 := lib.NewNH조회_질의값(lib.NH_TR_ETF_현재가_조회, lib.F임의_종목_ETF().G코드())

	응답_메시지 := F질의_NH(TR구분, 질의값)
	lib.F테스트_다름(t, 응답_메시지, nil)
	lib.F테스트_에러없음(t, 응답_메시지.G에러())
	lib.F테스트_같음(t, 응답_메시지.G길이(), 1)

	응답값 := lib.NewNH_ETF_현재가_조회_응답()
	lib.F테스트_에러없음(t, 응답_메시지.G값(0, 응답값))
	lib.F테스트_다름(t, 응답값, nil)
}

func TestTR실시간_서비스_등록_및_해지(t *testing.T) {
	if !lib.F한국증시_거래시간임() {
		t.SkipNow()
	}

	lib.F테스트_참임(t, F접속됨_NH()) // NH서버에 접속 되었는 지 확인.

	종목모음_코스피, 에러 := lib.F종목모음_코스피()
	lib.F테스트_에러없음(t, 에러)

	종목코드_모음 := make([]string, 0)
	for _, 종목 := range 종목모음_코스피 {
		종목코드_모음 = append(종목코드_모음, 종목.G코드())
		if len(종목코드_모음) > 20 {
			break
		}
	}

	// 실시간 정보 구독
	ch수신 := make(chan lib.I소켓_메시지, 10)
	에러 = F실시간_정보_구독_NH(ch수신, lib.NH_RT코스피_체결, 종목코드_모음)
	lib.F테스트_에러없음(t, 에러)

	// 실시간 정보 수신 확인
	for i := 0; i < 10; i++ {
		실시간_정보 := <-ch수신
		lib.F변수값_확인(실시간_정보)
	}

	// 실시간 정보 해지
	에러 = F실시간_정보_해지_NH(ch수신, lib.NH_RT코스피_체결, 종목코드_모음)
	lib.F테스트_에러없음(t, 에러)
}

func Test_ETF_틱_데이터_수집_NH(t *testing.T) {
	종목_모음, 에러 := lib.F종목모음_ETF()
	lib.F테스트_에러없음(t, 에러)

	defer func() {
		close(lib.F공통_종료_채널())
		lib.F공통_종료_채널_재설정()
	}()

	F실시간_데이터_수집_NH_ETF(lib.F2종목코드_모음(종목_모음))

	lib.F메모("실시간 데이터가 DB에저장되는 것을 확인할 것.")
}