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
	"github.com/ghts/lib"
	"github.com/go-mangos/mangos"
	"sync"
)

type s실시간_정보_구독_내역_저장소 struct {
	sync.Mutex
	중계_내역_저장소 map[chan lib.I소켓_메시지]mangos.Socket
	인덱스_소켓    []mangos.Socket
	인덱스_채널    [](chan lib.I소켓_메시지)
}

func (s *s실시간_정보_구독_내역_저장소) G소켓(채널 chan lib.I소켓_메시지) mangos.Socket {
	s.Lock()
	defer s.Unlock()

	return s.중계_내역_저장소[채널]
}

func (s *s실시간_정보_구독_내역_저장소) G중계_채널(소켓 mangos.Socket) (chan lib.I소켓_메시지, error) {
	s.Lock()
	defer s.Unlock()

	for ch수신, 소켓2 := range s.중계_내역_저장소 {
		if 소켓 == 소켓2 {
			return ch수신, nil
		}
	}

	return nil, lib.New에러("소켓에 해당되는 채널이 없음. %v", 소켓)
}

func (s *s실시간_정보_구독_내역_저장소) G소켓_모음() []mangos.Socket {
	s.Lock()
	defer s.Unlock()

	return s.인덱스_소켓
}

func (s *s실시간_정보_구독_내역_저장소) G채널_모음() [](chan lib.I소켓_메시지) {
	s.Lock()
	defer s.Unlock()

	return s.인덱스_채널
}

func (s *s실시간_정보_구독_내역_저장소) S추가(ch수신 chan lib.I소켓_메시지, 소켓 mangos.Socket) (에러 error) {
	defer lib.F에러패닉_처리(lib.S에러패닉_처리{M에러: &에러})

	// SUB소켓인지 확인
	옵션값, 에러 := 소켓.GetOption(mangos.OptionSubscribe)
	lib.F에러2패닉(에러)
	lib.F조건부_패닉(옵션값 == nil, "SUB소켓이 아닌 듯 함. 구독 옵션값이 nil임.")

	s.Lock()
	defer s.Unlock()

	s.중계_내역_저장소[ch수신] = 소켓
	s.s인덱스_재설정()

	return nil
}

func (s *s실시간_정보_구독_내역_저장소) S삭제(채널 chan lib.I소켓_메시지) (에러 error) {
	defer lib.F에러패닉_처리(lib.S에러패닉_처리{M에러: &에러})

	s.Lock()
	defer s.Unlock()

	delete(s.중계_내역_저장소, 채널)
	s.s인덱스_재설정()

	return nil
}

func (s *s실시간_정보_구독_내역_저장소) s인덱스_재설정() {
	i := 0
	s.인덱스_소켓 = make([]mangos.Socket, len(s.중계_내역_저장소))
	s.인덱스_채널 = make([]chan lib.I소켓_메시지, len(s.중계_내역_저장소))
	for 채널, 소켓 := range s.중계_내역_저장소 {
		s.인덱스_소켓[i] = 소켓
		s.인덱스_채널[i] = 채널
		i++
	}
}

func new실시간_정보_구독_내역_저장소() *s실시간_정보_구독_내역_저장소 {
	s := new(s실시간_정보_구독_내역_저장소)
	s.중계_내역_저장소 = make(map[chan lib.I소켓_메시지]mangos.Socket)
	s.인덱스_소켓 = make([]mangos.Socket, 0)
	s.인덱스_채널 = make([](chan lib.I소켓_메시지), 0)

	return s
}

type s대기_중_데이터_저장소 struct {
	sync.Mutex
	저장소 map[lib.I소켓_메시지](chan lib.I소켓_메시지)
}

func (s *s대기_중_데이터_저장소) S추가(메시지 lib.I소켓_메시지, ch수신 chan lib.I소켓_메시지) {
	s.Lock()
	defer s.Unlock()
	s.저장소[메시지] = ch수신
}

func (s *s대기_중_데이터_저장소) S재전송() {
	s.Lock()
	defer s.Unlock()

	for 메시지, ch수신 := range s.저장소 {
		s.s재전송_도우미(메시지, ch수신)
	}
}

func (s *s대기_중_데이터_저장소) s재전송_도우미(메시지 lib.I소켓_메시지, ch수신 chan lib.I소켓_메시지) {
	defer func() {
		// 채널이 이미 닫힌 경우 송신할 때 패닉이 발생함.
		// 그럴 경우 해당 메시지는 중계 대기열에서 삭제.
		delete(s.저장소, 메시지)
	}()

	select {
	case ch수신 <- 메시지:
		// 중계 성공한 메시지는 대기열에서 삭제.
		delete(s.저장소, 메시지)
	default:
		// 중계 실패. 저장소에 그대로 두고 추후 재전송 시도.
	}
}

func new대기_중_데이터_저장소() *s대기_중_데이터_저장소 {
	s := new(s대기_중_데이터_저장소)
	s.저장소 = make(map[lib.I소켓_메시지](chan lib.I소켓_메시지))

	return s
}