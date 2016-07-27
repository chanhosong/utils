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

	"testing"
)

func TestS일일가격정보_저장소_G단순이동평균(t *testing.T) {
	종목 := 공용.New종목("069500", "KODEX 200", 공용.P시장구분_ETF)
	초기_일자 := 공용.F2포맷된_일자_단순형("2006-01-02", "2010-01-01")
	기간 := 5
	기준일 := 초기_일자.AddDate(0, 0, 기간)

	// 가격정보 설정
	가격정보_저장소 := New일일가격정보_저장소()
	for i := 0; i <= 기간; i++ {
		시가 := 100 + 10*int64(i)
		종가 := 시가 + 10

		가격정보 := 공용.S일일_가격정보{M종목코드: 종목.G코드(),
			M일자: 초기_일자.AddDate(0, 0, i),
			M시가: 시가,
			M종가: 종가, M조정종가: 종가}

		가격정보_저장소.S추가(&가격정보)
	}

	단순이동평균_조정시가, 에러 := 가격정보_저장소.G단순이동평균(종목.G코드(), 기준일, 기간, 조정시가)
	공용.F테스트_에러없음(t, 에러)
	공용.F테스트_같음(t, 단순이동평균_조정시가, 130)

	단순이동평균_조정종가, 에러 := 가격정보_저장소.G단순이동평균(종목.G코드(), 기준일, 기간, 조정종가)
	공용.F테스트_에러없음(t, 에러)
	공용.F테스트_같음(t, 단순이동평균_조정종가, 140)
}
