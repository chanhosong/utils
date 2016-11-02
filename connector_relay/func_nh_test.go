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

	"testing"
)

func TestF질의TR_NH(t *testing.T) {
	lib.F메모("커넥터 모듈 실행해야 함.")
	lib.F대기(lib.P300밀리초)

	응답_메시지 := F질의_NH(nh.TR_ETF_현재가_조회, lib.F임의_종목_ETF())

	lib.F테스트_다름(t, 응답_메시지, nil)
	lib.F테스트_에러없음(t, 응답_메시지.G에러())
	lib.F테스트_같음(t, 응답_메시지.G길이(), 1)

	응답값 := nh.NewNH_ETF_현재가_조회_응답()
	lib.F테스트_에러없음(t, 응답_메시지.G값(0, 응답값))
	lib.F테스트_다름(t, 응답값, nil)
}

func TestTR실시간_서비스_등록_및_해지(t *testing.T) {
	if !lib.F한국증시_거래시간임() {
		return // 실시간 정보 테스트는 거래시간에만 하도록 함.
	}

	lib.F테스트_에러없음(t, f접속_확인())
	lib.F대기(lib.P300밀리초)
	변환_형식 := lib.F임의_변환형식()

	TR소켓, 에러 := lib.New소켓REQ(lib.P주소_NH_TR)
	lib.F테스트_에러없음(t, 에러)
	defer TR소켓.Close()

	구독_소켓_CBOR, 에러 := lib.New소켓SUB(lib.P주소_NH실시간_CBOR)
	lib.F테스트_에러없음(t, 에러)
	defer 구독_소켓_CBOR.Close()

	구독_소켓_JSON, 에러 := lib.New소켓SUB(lib.P주소_NH실시간_JSON)
	lib.F테스트_에러없음(t, 에러)
	defer 구독_소켓_JSON.Close()

	구독_소켓_MsgPack, 에러 := lib.New소켓SUB(lib.P주소_NH실시간_MsgPack)
	lib.F테스트_에러없음(t, 에러)
	defer 구독_소켓_MsgPack.Close()

	// 실시간 서비스 등록
	종목모음_코스피, 에러 := lib.F종목모음_코스피()
	lib.F테스트_에러없음(t, 에러)

	질의값 := NewNH실시간_정보_질의(RT코스피_체결, 종목모음_코스피)
	에러 = lib.F송신zmq(TR소켓, lib.P30초, 변환_형식, lib.TR실시간_정보_구독, 질의값)
	lib.F테스트_에러없음(t, 에러)

	응답 := lib.F수신zmq(TR소켓, lib.P30초)
	lib.F테스트_에러없음(t, 응답.G에러())
	lib.F테스트_같음(t, 응답.G길이(), 0)

	// 실시간 정보 수신 확인
	버퍼 := new(bytes.Buffer)
	for _, 종목 := range 질의값.M종목_모음 {
		버퍼.WriteString(종목.G코드())
	}

	전체_종목코드 := 버퍼.String()

	체결_정보 := new(NH체결)
	응답 = lib.F수신zmq(구독_소켓_CBOR, lib.P무기한)
	lib.F테스트_에러없음(t, 응답.G에러())
	lib.F테스트_에러없음(t, 응답.G값(0, 체결_정보))
	lib.F테스트_다름(t, strings.TrimSpace(체결_정보.M종목코드), "")
	lib.F테스트_참임(t, strings.Contains(전체_종목코드, 체결_정보.M종목코드))

	체결_정보 = new(NH체결)
	응답 = lib.F수신zmq(구독_소켓_JSON, lib.P무기한)
	lib.F테스트_에러없음(t, 응답.G에러())
	lib.F테스트_에러없음(t, 응답.G값(0, 체결_정보))
	lib.F테스트_다름(t, strings.TrimSpace(체결_정보.M종목코드), "")
	lib.F테스트_참임(t, strings.Contains(전체_종목코드, 체결_정보.M종목코드))

	체결_정보 = new(NH체결)
	응답 = lib.F수신zmq(구독_소켓_MsgPack, lib.P무기한)
	lib.F테스트_에러없음(t, 응답.G에러())
	lib.F테스트_에러없음(t, 응답.G값(0, 체결_정보))
	lib.F테스트_다름(t, strings.TrimSpace(체결_정보.M종목코드), "")
	lib.F테스트_참임(t, strings.Contains(전체_종목코드, 체결_정보.M종목코드))

	// 실시간 서비스 해지
	lib.F대기(lib.P300밀리초)
	에러 = lib.F송신zmq(TR소켓, lib.P30초, 변환_형식, lib.TR실시간_정보_해지, 질의값)
	lib.F테스트_에러없음(t, 에러)

	응답 = lib.F수신zmq(TR소켓, lib.P30초)
	lib.F테스트_에러없음(t, 응답.G에러())
	lib.F테스트_같음(t, 응답.G길이(), 0)
}
