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

package realtime_data_nh


import (
	"github.com/ghts/ghts_utils/connector_relay"
	"github.com/ghts/lib"
	"github.com/boltdb/bolt"
	"github.com/go-mangos/mangos"
	"time"
	"os"
	"sync"
)

type S틱_데이터 struct {
	M종목코드     string
	M시각       time.Time
	M매수_호가_모음 []int64
	M매수_잔량_모음 []int64
	M매도_호가_모음 []int64
	M매도_잔량_모음 []int64
	M현재가      int64
	NAV       float64
	M거래량      int64
}

func f틱_데이터_파일명(종목 *lib.S종목) string {
	return "tick_" + 종목.G코드() + "_" + time.Now().Format(lib.P일자_형식) + ".dat"
}

func f틱_데이터_수집_NH_ETF(SUB소켓 mangos.Socket, ch초기화 chan lib.T신호,
	ch수신 chan lib.I소켓_메시지, 종목코드_모음 []string) {
	defer lib.F에러패닉_처리(lib.S에러패닉_처리{M함수with패닉내역: func(r interface{}) { lib.New에러with출력(r) }})

	// NH 루틴 시작
	if 에러 := connector_relay.F초기화(); 에러 != nil {
		ch초기화 <- lib.P신호_초기화
		return
	}

	lib.F대기(lib.P500밀리초)

	에러 := connector_relay.F실시간_정보_구독_NH(ch수신, lib.NH_RT코스피_호가_잔량, 종목코드_모음)
	lib.F에러2패닉(에러)

	에러 = connector_relay.F실시간_정보_구독_NH(ch수신, lib.NH_RT코스피_체결, 종목코드_모음)
	lib.F에러2패닉(에러)

	에러 = connector_relay.F실시간_정보_구독_NH(ch수신, lib.NH_RT코스피_ETF_NAV, 종목코드_모음)
	lib.F에러2패닉(에러)

	NH호가_잔량 := lib.F자료형_문자열(lib.NH호가_잔량{})
	NH시간외_호가잔량 := lib.F자료형_문자열(lib.NH시간외_호가잔량{})
	NH예상_호가잔량 := lib.F자료형_문자열(lib.NH예상_호가잔량{})
	NH체결 := lib.F자료형_문자열(lib.NH체결{})
	NH_ETF_NAV := lib.F자료형_문자열(lib.NH_ETF_NAV{})
	NH업종_지수 := lib.F자료형_문자열(lib.NH업종_지수{})

	const db파일경로 = "nh_realtime_tick_data"
	const 버킷_호가_잔량 = "rt_bid_offer"
	const 버킷_시간외_호가_잔량 = "rt_offtime_bid_offer"
	const 버킷_예상_호가_잔량 = "rt_projected_bid_offer"
	const 버킷_체결 = "rt_trade"
	const 버킷_ETF_NAV = "rt_etf_nav"
	const 버킷_업종지수 = "rt_sector_index"

	db, 에러 := New데이터베이스(db파일경로)
	lib.F에러2패닉(에러)

	var 수신_메시지 lib.I소켓_메시지

	// 실시간 데이터 수신
	for {
		select {
		case <-lib.F공통_종료_채널():
			return
		case 수신_메시지 = <-ch수신:
			if 수신_메시지.G에러() != nil {
				continue
			}
		}

		lib.F조건부_패닉(수신_메시지.G길이() != 1, "예상하지 못한 메시지 길이. %v", 수신_메시지.G길이())

		var 질의 *S데이터베이스_질의
		var 에러 error

		switch 수신_메시지.G자료형_문자열(0) {
		case NH호가_잔량:
			s := new(lib.NH호가_잔량)
			에러 = 수신_메시지.G값(0, s)

			시각_바이트_모음, 에러 := s.M시각.MarshalBinary()
			lib.F에러2패닉(에러)

			질의 = new(S데이터베이스_질의)
			질의.M버킷명 = 버킷_호가_잔량
			질의.M키 = append([]byte(s.M종목코드), 시각_바이트_모음)
			질의.M값 =
		case NH시간외_호가잔량:
			s := new(lib.NH시간외_호가잔량)
			에러 = 수신_메시지.G값(0, s)

			시각_바이트_모음, 에러 := s.M시각.MarshalBinary()
			lib.F에러2패닉(에러)

			질의 = new(S데이터베이스_질의)
			질의.M버킷명 = 버킷_시간외_호가_잔량
			질의.M키 = append([]byte(s.M종목코드), 시각_바이트_모음)
			질의.M값 =
		case NH예상_호가잔량:
			s := new(lib.NH예상_호가잔량)
			에러 = 수신_메시지.G값(0, s)

			시각_바이트_모음, 에러 := s.M시각.MarshalBinary()
			lib.F에러2패닉(에러)

			질의 = new(S데이터베이스_질의)
			질의.M버킷명 = 버킷_예상_호가_잔량
			질의.M키 = append([]byte(s.M종목코드), 시각_바이트_모음)
			질의.M값 =
		case NH체결:
			s := new(lib.NH체결)
			에러 = 수신_메시지.G값(0, s)

			시각_바이트_모음, 에러 := s.M시각.MarshalBinary()
			lib.F에러2패닉(에러)

			질의 = new(S데이터베이스_질의)
			질의.M버킷명 = 버킷_체결
			질의.M키 = append([]byte(s.M종목코드), 시각_바이트_모음)
			질의.M값 =
		case NH_ETF_NAV:
			s := new(lib.NH_ETF_NAV)
			에러 = 수신_메시지.G값(0, s)

			시각_바이트_모음, 에러 := s.M시각.MarshalBinary()
			lib.F에러2패닉(에러)

			질의 = new(S데이터베이스_질의)
			질의.M버킷명 = 버킷_ETF_NAV
			질의.M키 = append([]byte(s.M종목코드), 시각_바이트_모음)
			질의.M값 =
		case NH업종_지수:
			s := new(lib.NH업종_지수)
			에러 = 수신_메시지.G값(0, s)

			시각_바이트_모음, 에러 := s.M시각.MarshalBinary()
			lib.F에러2패닉(에러)

			질의 = new(S데이터베이스_질의)
			질의.M버킷명 = 버킷_업종지수
			질의.M키 = append([]byte(s.M업종_코드), 시각_바이트_모음)
			질의.M값 =
		default:
			lib.F패닉("예상하지 못한 자료형. %v", 수신_메시지.G자료형_문자열(0))
		}

		// DB에 수신값 저장
		if 에러 = db.S업데이트(질의); 에러 != nil {
			lib.F에러_출력(에러)
			continue
		}
	}
}

type S데이터베이스_질의 struct {
	M버킷명 []byte
	M키 []byte
	M값 []byte
}

type S데이터베이스_회신 struct {
	M에러 error
	M값 []byte
}

// Bolt DB 사용
type I데이터베이스 interface {
	G질의(질의 *S데이터베이스_질의) *S데이터베이스_회신
	S업데이트(질의 *S데이터베이스_질의) error
	S삭제(질의 *S데이터베이스_질의) error
}

type s데이터베이스 struct {
	sync.RWMutex
	db *bolt.DB
}

func (s *s데이터베이스) G질의(질의 *S데이터베이스_질의) *S데이터베이스_회신 {
	s.RLock()   // 동시 읽기가 가능함.
	defer s.RUnlock()

	ch회신 := make(chan *S데이터베이스_회신)

	s.db.View(func(tx *bolt.Tx) (에러 error) {
		defer lib.F에러패닉_처리(lib.S에러패닉_처리{
			M에러: &에러,
			M함수: func() { ch회신 <- &(S데이터베이스_회신{M에러: 에러, M값: nil}) }
		})

		바이트_모음 := tx.Bucket(질의.M버킷명).Get(질의.M키)
		ch회신 <- &(S데이터베이스_회신{M에러: nil, M값: 바이트_모음})
		return nil
	})

	return <-ch회신
}

func (s *s데이터베이스) S업데이트(질의 *S데이터베이스_질의) error {
	s.Lock()   // 동시 작업 불가.
	defer s.Unlock()

	return s.db.Update(func(tx *bolt.Tx) (에러 error) {
		defer lib.F에러패닉_처리(lib.S에러패닉_처리{M에러: &에러})

		버킷, 에러 := tx.CreateBucketIfNotExists(질의.M버킷명)
		lib.F에러2패닉(에러)

		return 버킷.Put(질의.M키, 질의.M값)
	})
}

func (s *s데이터베이스) S삭제(질의 *S데이터베이스_질의) error {
	s.Lock()   // 동시 작업 불가.
	defer s.Unlock()

	return s.db.Update(func(tx *bolt.Tx) (에러 error) {
		defer lib.F에러패닉_처리(lib.S에러패닉_처리{M에러: &에러})

		return tx.Bucket(질의.M버킷명).Delete(질의.M키)
	})
}

func New데이터베이스(DB파일경로 string) (db *s데이터베이스, 에러 error) {
	defer lib.F에러패닉_처리(lib.S에러패닉_처리{
		M에러: &에러,
		M함수: func() { db = nil }})

	1개의 인스턴스만 생성하도록 할 것.

	s := new(s데이터베이스)
	s.db, 에러 = bolt.Open(db파일경로, os.ModeExclusive, nil)

	return s, 에러
}
