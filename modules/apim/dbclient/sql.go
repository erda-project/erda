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
