//go:build !go1.19

package qtls

var _ int = "The version of quic-go you're using can't be built using outdated Go versions."
