package main

import (
	"testing"
	"time"
)

func mk(t *testing.T) *Store {
	s, err := openStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestRoutineLead(t *testing.T) {
	s := mk(t)
	day := time.Date(2026, 11, 4, 0, 0, 0, 0, time.Local) // quarta
	s.d.Routines = []Routine{
		{ID: "a", Title: "Reunião Trabalho", Time: "10:00", Fixed: true, Days: []int{1, 2, 3, 4, 5}},
		{ID: "b", Title: "Academia", Time: "18:00", Date: "2026-11-04"},
		{ID: "c", Title: "Outro dia", Time: "10:00", Date: "2026-11-05"},
	}
	at := func(h, m int) time.Time { return day.Add(time.Duration(h)*time.Hour + time.Duration(m)*time.Minute) }
	if n, _ := s.check(at(9, 49)); len(n) != 0 {
		t.Fatalf("cedo demais: %v", n)
	}
	n, _ := s.check(at(9, 50))
	if len(n) != 1 || n[0].Title != "⏰ Reunião Trabalho" {
		t.Fatalf("esperava aviso 9:50: %v", n)
	}
	t.Log(n[0].Msg)
	if n, _ := s.check(at(9, 55)); len(n) != 0 {
		t.Fatalf("repetiu: %v", n)
	}
	if n, _ := s.check(at(10, 1)); len(n) != 0 {
		t.Fatalf("depois: %v", n)
	}
	// concluída antes do aviso não avisa
	s.d.Done["2026-11-04"] = []string{"b"}
	if n, _ := s.check(at(17, 52)); len(n) != 0 {
		t.Fatalf("feito não deve avisar: %v", n)
	}
}

func TestTaskMilestones(t *testing.T) {
	s := mk(t)
	created := time.Date(2026, 11, 1, 9, 0, 0, 0, time.Local)
	due := time.Date(2026, 11, 5, 18, 0, 0, 0, time.Local)
	tk := Task{ID: "t1", Name: "Trabalho de Cálculo", Type: "Faculdade", Start: "2026-11-03T14:00", Due: due.Format(stampFmt)}
	s.d.Tasks = []Task{tk}
	s.markPassed(tk, created)
	var got []string
	for tm := created; tm.Before(due.Add(8 * time.Hour)); tm = tm.Add(15 * time.Second) {
		for _, n := range func() []Note { n, _ := s.check(tm); return n }() {
			got = append(got, tm.Format("02 15:04")+" "+n.Title)
		}
	}
	for _, g := range got {
		t.Log(g)
	}
	want := []string{"03 14:00 🚀 Hora de começar", "03 18:00 📌 Faltam 2 dias", "04 18:00 📌 Faltam 1 dia", "05 06:00 📌 Faltam 12 horas", "05 17:00 📌 Faltam 1 hora", "05 18:00 ⚠️ Prazo de entrega chegou"}
	if len(got) != len(want) {
		t.Fatalf("got %d", len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%d: %q != %q", i, got[i], want[i])
		}
	}
}

func TestTaskCreatedLate(t *testing.T) {
	s := mk(t)
	now := time.Date(2026, 11, 4, 10, 0, 0, 0, time.Local)
	tk := Task{ID: "t2", Name: "Relatório", Due: "2026-11-05T08:00"} // 22h depois
	s.d.Tasks = []Task{tk}
	s.markPassed(tk, now)
	if n, _ := s.check(now); len(n) != 0 {
		t.Fatalf("não devia avisar 2d/1d: %v", n)
	}
	// PC desligado no marco de 12h; liga às 23:00 (9h antes) -> 1 aviso só
	n, _ := s.check(time.Date(2026, 11, 4, 23, 0, 0, 0, time.Local))
	if len(n) != 1 {
		t.Fatalf("esperava 1: %v", n)
	}
	t.Log(n[0].Title, "|", n[0].Msg)
}
