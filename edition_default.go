//go:build !prf

package main

// Edição padrão: sem conteúdo extra. Outras edições (ex.: -tags prf) trazem
// catálogos e tema próprios, aplicados só na primeira execução.
func editionSeed(s *Store) bool { return false }
