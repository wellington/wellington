package spritewell

type Spriter interface {
	Sprite() *SafeImageMap
}

type Imager interface {
	Image() *SafeImageMap
}
