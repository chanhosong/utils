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

package main

import (
	"github.com/ghts/lib"
)

func fTX실행(sql문 string, 인자_모음 ...interface{}) (에러 error) {
	defer lib.F에러패닉_처리(lib.S에러패닉_처리{M에러 : &에러})

	db, 에러 := fMySQL_DB()
	lib.F에러2패닉(에러)

	tx, 에러 := db.Begin()
	lib.F에러2패닉(에러)

	stmt, 에러 := tx.Prepare(sql문)
	lib.F에러2패닉(에러)

	_, 에러 = stmt.Exec(인자_모음...)
	lib.F에러2패닉(에러)

	lib.F에러2패닉(tx.Commit())

	return 에러
}

func f정수값_질의(sql문 string, 인자_모음 ...interface{}) (값 int, 에러 error) {
	defer lib.F에러패닉_처리(lib.S에러패닉_처리{
		M에러 : &에러,
		M함수 : func() { 값 = 0 }})

	db, 에러 := fMySQL_DB()
	lib.F에러2패닉(에러)

	lib.F에러2패닉(db.QueryRow(sql문, 인자_모음...).Scan(&값))

	return 값, 에러
}