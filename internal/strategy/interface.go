package strategy

type Strategy interface {
    Execute() error
    GetStatus() string
}
