package shop

type MenuItem interface {
	Image() bool
	Title() string
	ImageName() string
}
