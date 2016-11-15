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
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	lib.F테스트_모드_시작()
	defer lib.F테스트_모드_종료()
	defer close(lib.F공통_종료_채널())

	// 증권사 API 커넥터 실행
	const 증권사API_커넥터_실행파일_경로 = ""

	ch실행결과 := make(chan lib.I채널_메시지, 1)
	lib.F외부_프로세스_실행(ch실행결과, lib.P무기한, 증권사API_커넥터_실행파일_경로)

	실행결과 := <-ch실행결과
	lib.F에러2패닉(실행결과.G에러())

	var 프로세스ID int
	if 실행결과.G값(0).(lib.T신호) == lib.P신호_초기화 {
		프로세스ID = 실행결과.G값(1).(int)
	}

	테스트_실행결과 := m.Run()

	에러 := lib.F프로세스_종료by프로세스ID(프로세스ID)
	lib.F에러2패닉(에러)

	os.Exit(테스트_실행결과)
}