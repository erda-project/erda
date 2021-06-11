// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package dbclient

const DistinctAssetIDFromAccess = `
SELECT SQL_CALC_FOUND_ROWS asset_id
FROM dice_api_access
where org_id = ?
	AND ('' = ? OR asset_id = ? OR asset_name LIKE ?)
	AND (asset_id IN (?))
ORDER BY updated_at DESC
LIMIT ? OFFSET ?
`

const SelectFoundRows = `
SELECT FOUND_ROWS()
`
