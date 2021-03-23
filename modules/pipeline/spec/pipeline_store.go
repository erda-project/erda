package spec

type StoreType string

var (
	StoreTypeOSS             StoreType = "OSS"
	StoreTypeNFS             StoreType = "NFS"
	StoreTypeDiceVolumeNFS   StoreType = "dice-nfs-volume"
	StoreTypeDiceVolumeLocal StoreType = "dice-local-volume"
	StoreTypeDiceVolumeFake  StoreType = "dice-fake-volume"
	StoreTypeDiceCacheNFS    StoreType = "dice-cache-nfs-volume"
)

const (
	StoreTypeNFSProto = "file://"
)
