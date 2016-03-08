package reminders

import (
	"fmt"
	"github.com/jirkadanek/onlinea/bc"
	"github.com/jirkadanek/onlinea/ismu/api"
	"github.com/jirkadanek/onlinea/mailing"
	"github.com/jirkadanek/onlinea/secrets"
	"io/ioutil"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
)

//import (
//	"github.com/jirkadanek/onlinea/bc"
//	"io/ioutil"
//"strings"
//	"github.com/jirkadanek/onlinea/mailing"
//)
//
//func PointsForStudent(block map[string]string) map[string]int {
//
//}
//
//func Program() (string, string) {
//	program, err := ioutil.ReadFile("first.bc")
//	names := blocknames(program)
//	return program, names
//}
//

func blockName(v string) string {
	p := "ma_body_"
	if strings.HasPrefix(v, p) {
		return v[len(p):len(v)]
	}
	return v
}

//
//func variables(blocksstudentpoints, student) map[string]string {
//		vars := make(map[string]string)
//		for block, studentpoints := range blockstudentpoints {
//			vars[block], ma_body := studentpoints[student]
//			vars["ma_body_" + block] = 0
//			if ma_body {
//				vars["ma_body_" + block] = 1
//			}
//		}
//
//	}
//}
//
//
//
//
//func blockNamesToFetch(variables map[string]string) []string {
//	blocknames := getBlockNames()
//	blocks := []string{"asumtotal", "asumdisk"}
//	for _, variable := range variables {
//		if _, found := blocknames[variable]; found {
//			blocks = append(blocks, variable)
//		}
//	}
//	return blocks
//}
//
//func fetchBlockData() {
//		blockstudentpoints := make(map[string]map[string]int)
//	for _, block := range nblocks {
//		blockstudentpoints[block] = PointsForStudent(block)
//	}
//}
//
//func createReminders() map[string]mailing.ReminderData {
//	reminders := make(map[string]mailing.ReminderData, 0)
//	for _, student := range students {
//		reminders[student] = mailing.ReminderData{Email: student + "@mail.muni.cz"}
//	}
//
//	for _, student := range students {
//		reminder := &reminders[student]
//
//		vars := variables(blockstudentpoints, student)
//		output, err := bc.Run(program, vars)
//		reminders[student].DeadlinePrintout = output
//
//		completed := strings.Index(output, "@N") == -1
//
//		reminder.DeadlineCompleted = completed
//
//		reminder.Discussion = blockstudentpoints["asumdiscussion"][student]
//		reminder.Total = blockstudentpoints["asumatotal"][student]
//		reminder.DeadlinesMissed = blocksstudentpoints["missed"][student]
//	}
//	return reminders
//}
//
type WeeklyProgressReminders struct {
	client     *api.Client
	parameters api.Parameters
	name       string
	number     int
	program    string
	variables  []string
	students   []api.CourseStudent
	notebooks  map[string][]api.Notebook
	//	reminders map[string]mailing.ReminderData
	Err error
}

func NewWeeklyProgressReminders(client *api.Client, parameters api.Parameters) WeeklyProgressReminders {
	return WeeklyProgressReminders{client: client, parameters: parameters}
}

func (r *WeeklyProgressReminders) Failed() bool {
	return r.Err != nil
}

func (r *WeeklyProgressReminders) Deadline(name string, number int, skript string) {
	if r.Failed() {
		return
	}
	r.name = name
	r.number = number
	var program []byte
	program, r.Err = ioutil.ReadFile(skript)
	r.program = string(program)
	if r.Failed() {
		return
	}
	r.variables = bc.Variables(r.program)
}

func (r *WeeklyProgressReminders) GetStudents() {
	if r.Failed() {
		return
	}
	r.students, r.Err = r.client.GetCourseStudents(r.parameters, api.Zaregistrovani)
}

func (r *WeeklyProgressReminders) GetNotebooks() {
	if r.Failed() {
		return
	}

	required := make(map[string]struct{}, 0)
	for _, b := range []string{"asumadisk", "asumatotal", "asumamiss"} {
		required[b] = struct{}{}
	}
	for _, v := range r.variables {
		required[blockName(v)] = struct{}{}
	}

	var list []api.NotebookInfo
	list, r.Err = r.client.GetNotebookList(r.parameters)
	if r.Failed() {
		return
	}

	r.notebooks = make(map[string][]api.Notebook)
	for _, l := range list {
		if _, found := required[l.ZKRATKA]; found {
			r.notebooks[l.ZKRATKA], r.Err = r.client.GetNotebook(r.parameters, l.ZKRATKA, []string{})
			if r.Failed() {
				return
			}
		}
	}
}

func (r *WeeklyProgressReminders) Reminders() []mailing.ReminderData {
	rs := make([]mailing.ReminderData, len(r.students))
	for i, s := range r.students {
		defs := make(map[string]string)
		for k, _ := range r.notebooks {
			p := points(obsah(r.notebooks, k, s.UCO))

			defs[k] = fmt.Sprintf("%f", p)

			defs["ma_body_"+k] = "0"
			if p != 0 {
				defs["ma_body_"+k] = "1"
			}
		}
		printout, err := bc.Run(r.program, defs)
		if err != nil {
			log.Fatal(err)
		}

		rs[i] = mailing.ReminderData{
			Email:             s.UCO + "@mail.muni.cz",
			FullName:          s.CELE_JMENO,
			DeadlineName:      r.name,
			DeadlineNumber:    r.number,
			DeadlineCompleted: completed(printout),
			DeadlinePrintout:  printout,
			DeadlinesMissed:   int(math.Floor(points(obsah(r.notebooks, "asumamiss", s.UCO)))),
			Discussion:        points(obsah(r.notebooks, "asumadisk", s.UCO)),
			Total:             points(obsah(r.notebooks, "asumatotal", s.UCO)),
		}
	}
	return rs
}

func (r *WeeklyProgressReminders) Perform(name string, number int, skript string) []mailing.ReminderData {
	r.Deadline(name, number, skript)
	r.GetStudents()
	r.GetNotebooks()
	return r.Reminders()
}

func points(s string) float64 {
	sum := 0.0
	//regexp.MustCompile(`\*(\d+(?:[\.,]\d+))`)
	r := regexp.MustCompile(`\*(-?\d+(?:[\.,]\d+)?)`)
	for _, m := range r.FindAllStringSubmatch(s, -1) {
		if len(m) == 2 {
			f, err := strconv.ParseFloat(strings.Replace(m[1], ",", ".", -1), 64)
			if err != nil {
				log.Fatal(err)
			}
			sum += f
		}
	}
	return sum
}

func completed(s string) bool {
	return strings.Index(s, "@N") == -1
}

func obsah(bloky map[string][]api.Notebook, zkratka, uco string) string {
	var record string
	for _, rec := range bloky[zkratka] {
		if rec.UCO == uco {
			record = rec.OBSAH
		}
	}
	return record
}

func WetRun() {
	parameters := api.Parameters{Fakulta: "1441", Kod: "ONLINE_A"}
	client := api.NewClient(secrets.APIKEY, nil)
	pc := NewWeeklyProgressReminders(client, parameters)
	reminders := pc.Perform("Deadline Assignment 1", 1, "first.bc")
	if pc.Failed() {
		log.Fatal(pc.Err)
	}
	log.Printf("%+v", reminders)
}

//func (self *Reminders) loadProgram() {
//
//}
//
//func (self *Reminders) buildReminders(client api.client) {
//	blocknames := client.blockNamesToFetch(variables)
//	blockstudentpoints := fetchBlockData(blocknames)
//	reminders := createreminders(blockstudentpoints)
//	return reminders
//}
