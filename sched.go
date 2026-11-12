package main

import (
	"fmt"
	"time"
)

type milestone struct {
	key string
	d   time.Duration
}

// Avisos antes da entrega: 2 dias, 1 dia, 12 horas e 1 hora.
var milestones = []milestone{
	{"48h", 48 * time.Hour},
	{"24h", 24 * time.Hour},
	{"12h", 12 * time.Hour},
	{"1h", time.Hour},
}

type Note struct{ Title, Msg string }

var weekdaysPT = []string{"dom", "seg", "ter", "qua", "qui", "sex", "sáb"}

func fmtWhen(t, now time.Time) string {
	hm := t.Format(clockFmt)
	y1, m1, d1 := now.Date()
	y2, m2, d2 := t.Date()
	switch {
	case y1 == y2 && m1 == m2 && d1 == d2:
		return "hoje às " + hm
	case now.AddDate(0, 0, 1).Format(dateFmt) == t.Format(dateFmt):
		return "amanhã às " + hm
	default:
		return fmt.Sprintf("%s %s às %s", weekdaysPT[t.Weekday()], t.Format("02/01"), hm)
	}
}

func fmtLeft(d time.Duration) string {
	if d < time.Minute {
		return "menos de 1 minuto"
	}
	mins := int(d.Round(time.Minute).Minutes())
	days, hours, m := mins/1440, (mins%1440)/60, mins%60
	plural := func(n int, one, many string) string {
		if n == 1 {
			return "1 " + one
		}
		return fmt.Sprintf("%d %s", n, many)
	}
	switch {
	case days > 0 && hours > 0:
		return plural(days, "dia", "dias") + " e " + plural(hours, "hora", "horas")
	case days > 0:
		return plural(days, "dia", "dias")
	case hours > 0 && m > 0 && hours < 3:
		return plural(hours, "hora", "horas") + " e " + plural(m, "minuto", "minutos")
	case hours > 0:
		return plural(hours, "hora", "horas")
	default:
		return plural(m, "minuto", "minutos")
	}
}

// markPassed marca em silêncio os avisos que já passaram (ao criar/editar uma tarefa),
// para não disparar "faltam 2 dias" numa tarefa criada com 1 dia de prazo.
func (s *Store) markPassed(t Task, now time.Time) {
	today := now.Format(dateFmt)
	if due, ok := parseStamp(t.Due); ok {
		for _, m := range milestones {
			if !now.Before(due.Add(-m.d)) {
				s.d.Sent["t|"+t.ID+"|"+m.key] = today
			}
		}
		if !now.Before(due) {
			s.d.Sent["t|"+t.ID+"|due"] = today
		}
	}
	if st, ok := parseStamp(t.Start); ok && !now.Before(st) {
		s.d.Sent["t|"+t.ID+"|start"] = today
	}
}

// check roda a cada poucos segundos e devolve as notificações a disparar.
func (s *Store) check(now time.Time) (notes []Note, changed bool) {
	today := now.Format(dateFmt)
	mark := func(k string) bool {
		if _, ok := s.d.Sent[k]; ok {
			return false
		}
		s.d.Sent[k] = today
		changed = true
		return true
	}

	// Rotina do dia: aviso X minutos antes
	lead := time.Duration(s.d.Settings.Lead) * time.Minute
	for _, r := range s.d.Routines {
		if !r.occursOn(now) || s.isDone(today, r.ID) {
			continue
		}
		at, ok := atClock(now, r.Time)
		if !ok {
			continue
		}
		if !now.Before(at.Add(-lead)) && now.Before(at) {
			if mark("r|" + today + "|" + r.ID) {
				left := at.Sub(now)
				msg := fmt.Sprintf("Começa às %s (em %s).", r.Time, fmtLeft(left))
				if c := s.catalog(r.Area); c != nil {
					msg = c.Emoji + " " + c.Name + " · " + msg
				}
				if r.Note != "" {
					msg += "\n" + r.Note
				}
				notes = append(notes, Note{"⏰ " + r.Title, msg})
			}
		}
	}

	// Tarefas
	for _, t := range s.d.Tasks {
		due, ok := parseStamp(t.Due)
		if !ok {
			continue
		}
		label := t.Name
		if c := s.catalog(t.Area); c != nil {
			label = t.Name + " (" + c.Emoji + " " + c.Name + ")"
		} else if t.Type != "" {
			label = t.Name + " (" + t.Type + ")"
		}
		if st, ok := parseStamp(t.Start); ok && !now.Before(st) {
			k := "t|" + t.ID + "|start"
			if _, sent := s.d.Sent[k]; !sent {
				mark(k)
				if now.Sub(st) < 30*time.Minute && now.Before(due) {
					notes = append(notes, Note{"🚀 Hora de começar",
						fmt.Sprintf("%s\nEntrega %s. Bora, sem enrolar!", label, fmtWhen(due, now))})
				}
			}
		}
		if now.Before(due) {
			fire := false
			for _, m := range milestones {
				if !now.Before(due.Add(-m.d)) && mark("t|"+t.ID+"|"+m.key) {
					fire = true
				}
			}
			if fire {
				notes = append(notes, Note{"📌 Faltam " + fmtLeft(due.Sub(now)),
					fmt.Sprintf("%s\nEntrega %s.", label, fmtWhen(due, now))})
			}
		} else if mark("t|" + t.ID + "|due") {
			if now.Sub(due) < 6*time.Hour {
				notes = append(notes, Note{"⚠️ Prazo de entrega chegou",
					label + "\nSe já entregou, marque como concluída."})
			}
		}
	}

	if changed {
		s.prune(now)
	}
	return notes, changed
}
