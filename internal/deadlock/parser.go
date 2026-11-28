package deadlock

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"slices"
	"strconv"
	"strings"
)

// Pattern keys for deterministic deadlock classification.
const (
	PatternReverseOrder     = "reverse_order"
	PatternBookmarkLookup   = "bookmark_lookup"
	PatternLockUpgrade      = "lock_upgrade"
	PatternParallelExchange = "parallel_exchange"
	PatternGenericCycle     = "generic_cycle"
)

type ExecutionFrame struct {
	ProcName  string `json:"proc_name"`
	Line      string `json:"line"`
	SQLHandle string `json:"sql_handle"`
	SQL       string `json:"sql"`
}

type Process struct {
	ID              string           `json:"id"`
	IsVictim        bool             `json:"is_victim"`
	SPID            string           `json:"spid"`
	ECID            string           `json:"ecid"`
	Status          string           `json:"status"`
	WaitTimeMs      int64            `json:"wait_time_ms"`
	WaitResource    string           `json:"wait_resource"`
	LockMode        string           `json:"lock_mode"`
	LogUsed         int64            `json:"log_used"`
	Priority        string           `json:"priority"`
	TranCount       string           `json:"tran_count"`
	TransactionName string           `json:"transaction_name"`
	IsolationLevel  string           `json:"isolation_level"`
	ClientApp       string           `json:"client_app"`
	HostName        string           `json:"host_name"`
	LoginName       string           `json:"login_name"`
	CurrentDB       string           `json:"current_db"`
	CurrentDBName   string           `json:"current_db_name"`
	Frames          []ExecutionFrame `json:"frames"`
	InputBuf        string           `json:"input_buf"`
	ActiveSQL       string           `json:"active_sql"`
}

type ResourceParticipant struct {
	ProcessID   string `json:"process_id"`
	Mode        string `json:"mode"`
	RequestType string `json:"request_type"`
}

type Resource struct {
	ID          string                `json:"id"`
	Type        string                `json:"type"`
	DBID        string                `json:"dbid"`
	Database    string                `json:"database"`
	ObjectName  string                `json:"object_name"`
	ShortObject string                `json:"short_object"`
	IndexName   string                `json:"index_name"`
	HoBTID      string                `json:"hobt_id"`
	FileID      string                `json:"file_id"`
	PageID      string                `json:"page_id"`
	Slot        string                `json:"slot"`
	Mode        string                `json:"mode"`
	Owners      []ResourceParticipant `json:"owners"`
	Waiters     []ResourceParticipant `json:"waiters"`
}

type CycleStep struct {
	Kind string `json:"kind"` // "hold" | "wait" | "victim"
	Text string `json:"text"`
}

type Diagnosis struct {
	Pattern         string      `json:"pattern"`
	PatternLabel    string      `json:"pattern_label"`
	Severity        string      `json:"severity"` // "error" | "warning" | "purple" | "info"
	Summary         string      `json:"summary"`
	Recommendations []string    `json:"recommendations"`
	Steps           []CycleStep `json:"steps"`
}

type ParsedDeadlock struct {
	XMLHash     string     `json:"xml_hash"`
	Fingerprint string     `json:"fingerprint"`
	VictimSPID  string     `json:"victim_spid"`
	Databases   []string   `json:"databases"`
	MaxWaitMs   int64      `json:"max_wait_ms"`
	Processes   []Process  `json:"processes"`
	Resources   []Resource `json:"resources"`
	Diagnosis   Diagnosis  `json:"diagnosis"`
}

// HashXML computes a stable SHA-256 hex digest of a normalized deadlock XML payload.
func HashXML(rawXML string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(rawXML)))
	return hex.EncodeToString(sum[:])
}

// indexASCIIFold finds the first occurrence of ASCII lowercase needle in s using ASCII-only
// case folding so returned byte indices always match s directly (even with multi-byte Unicode).
func indexASCIIFold(s, lowerNeedle string) int {
	n := len(lowerNeedle)
	if n == 0 {
		return 0
	}
	if n > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-n; i++ {
		match := true
		for j := 0; j < n; j++ {
			c := s[i+j]
			if c >= 'A' && c <= 'Z' {
				c += 'a' - 'A'
			}
			if c != lowerNeedle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// ExtractDeadlockElement strips optional outer <event>/<deadlock-list> wrappers so stored XML
// starts cleanly at <deadlock> and can be exported directly as an SSMS .xdl file.
func ExtractDeadlockElement(rawXML string) string {
	trimmed := strings.TrimSpace(rawXML)
	if trimmed == "" {
		return ""
	}
	if indexASCIIFold(trimmed, "<deadlock") < 0 && indexASCIIFold(trimmed, "&lt;deadlock") >= 0 {
		trimmed = html.UnescapeString(trimmed)
	}
	start := indexASCIIFold(trimmed, "<deadlock")
	if start < 0 {
		return trimmed
	}
	// Avoid matching <deadlock-list
	for start >= 0 {
		after := start + len("<deadlock")
		if after < len(trimmed) && (trimmed[after] == '>' || trimmed[after] == ' ' || trimmed[after] == '\t' || trimmed[after] == '\n' || trimmed[after] == '\r') {
			break
		}
		next := indexASCIIFold(trimmed[after:], "<deadlock")
		if next < 0 {
			start = -1
			break
		}
		start = after + next
	}
	if start < 0 {
		return trimmed
	}
	endTag := "</deadlock>"
	relEnd := indexASCIIFold(trimmed[start:], endTag)
	if relEnd <= 0 {
		return trimmed
	}
	return strings.TrimSpace(trimmed[start : start+relEnd+len(endTag)])
}

type rawDeadlock struct {
	VictimList   rawVictimList   `xml:"victim-list"`
	ProcessList  rawProcessList  `xml:"process-list"`
	ResourceList rawResourceList `xml:"resource-list"`
}

type rawVictimList struct {
	Victims []struct {
		ID string `xml:"id,attr"`
	} `xml:"victimProcess"`
}

type rawProcessList struct {
	Processes []rawProcess `xml:"process"`
}

type rawProcess struct {
	ID              string `xml:"id,attr"`
	SPID            string `xml:"spid,attr"`
	ECID            string `xml:"ecid,attr"`
	Status          string `xml:"status,attr"`
	WaitTime        string `xml:"waittime,attr"`
	WaitResource    string `xml:"waitresource,attr"`
	LockMode        string `xml:"lockMode,attr"`
	LogUsed         string `xml:"logused,attr"`
	Priority        string `xml:"priority,attr"`
	TranCount       string `xml:"trancount,attr"`
	TransactionName string `xml:"transactionname,attr"`
	IsolationLevel  string `xml:"isolationlevel,attr"`
	ClientApp       string `xml:"clientapp,attr"`
	HostName        string `xml:"hostname,attr"`
	LoginName       string `xml:"loginname,attr"`
	CurrentDB       string `xml:"currentdb,attr"`
	CurrentDBName   string `xml:"currentdbname,attr"`
	ExecutionStack  struct {
		Frames []struct {
			ProcName  string `xml:"procname,attr"`
			Line      string `xml:"line,attr"`
			SQLHandle string `xml:"sqlhandle,attr"`
			Text      string `xml:",chardata"`
		} `xml:"frame"`
	} `xml:"executionStack"`
	InputBuf string `xml:"inputbuf"`
}

type rawResourceList struct {
	Items []rawResourceItem
}

type rawResourceItem struct {
	Type    string
	Attrs   map[string]string
	Owners  []ResourceParticipant
	Waiters []ResourceParticipant
}

func (rl *rawResourceList) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for {
		tok, err := d.Token()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			item, err := decodeResourceItem(d, t)
			if err != nil {
				return err
			}
			rl.Items = append(rl.Items, item)
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return nil
			}
		}
	}
}

func decodeResourceItem(d *xml.Decoder, start xml.StartElement) (rawResourceItem, error) {
	item := rawResourceItem{
		Type:  start.Name.Local,
		Attrs: make(map[string]string, len(start.Attr)),
	}
	for _, a := range start.Attr {
		item.Attrs[a.Name.Local] = a.Value
	}

	var section string
	depth := 1
	for depth > 0 {
		tok, err := d.Token()
		if err != nil {
			return item, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			switch t.Name.Local {
			case "owner-list":
				section = "owners"
			case "waiter-list":
				section = "waiters"
			case "owner":
				if section == "owners" || section == "" {
					item.Owners = append(item.Owners, decodeParticipant(t, ""))
				}
			case "waiter":
				if section == "waiters" || section == "" {
					item.Waiters = append(item.Waiters, decodeParticipant(t, "wait"))
				}
			}
		case xml.EndElement:
			depth--
			if t.Name.Local == "owner-list" || t.Name.Local == "waiter-list" {
				section = ""
			}
		}
	}
	return item, nil
}

func decodeParticipant(start xml.StartElement, defaultReqType string) ResourceParticipant {
	p := ResourceParticipant{RequestType: defaultReqType}
	for _, a := range start.Attr {
		switch a.Name.Local {
		case "id":
			p.ProcessID = a.Value
		case "mode":
			p.Mode = a.Value
		case "requestType":
			if a.Value != "" {
				p.RequestType = a.Value
			}
		}
	}
	return p
}

// ParseXML parses a SQL Server xml_deadlock_report payload into structured processes,
// contested resources, fingerprint, and automated root-cause diagnosis.
func ParseXML(rawXML string) (*ParsedDeadlock, error) {
	cleanXML := ExtractDeadlockElement(rawXML)
	if cleanXML == "" {
		return nil, fmt.Errorf("empty deadlock XML")
	}

	var rd rawDeadlock
	if err := xml.Unmarshal([]byte(cleanXML), &rd); err != nil {
		return nil, fmt.Errorf("unmarshal deadlock xml: %w", err)
	}

	victimIDs := make(map[string]struct{}, len(rd.VictimList.Victims))
	for _, v := range rd.VictimList.Victims {
		if v.ID != "" {
			victimIDs[v.ID] = struct{}{}
		}
	}

	dbSet := make(map[string]struct{})
	processes := make([]Process, 0, len(rd.ProcessList.Processes))
	var maxWaitMs int64
	var victimSPID string

	for _, rp := range rd.ProcessList.Processes {
		_, isVictim := victimIDs[rp.ID]
		waitMs, _ := strconv.ParseInt(strings.TrimSpace(rp.WaitTime), 10, 64)
		logUsed, _ := strconv.ParseInt(strings.TrimSpace(rp.LogUsed), 10, 64)
		if waitMs > maxWaitMs {
			maxWaitMs = waitMs
		}
		if isVictim && victimSPID == "" {
			victimSPID = rp.SPID
		}

		frames := make([]ExecutionFrame, 0, len(rp.ExecutionStack.Frames))
		for _, f := range rp.ExecutionStack.Frames {
			procName := strings.TrimSpace(f.ProcName)
			if procName == "" {
				procName = "adhoc"
			}
			frames = append(frames, ExecutionFrame{
				ProcName:  procName,
				Line:      strings.TrimSpace(f.Line),
				SQLHandle: strings.TrimSpace(f.SQLHandle),
				SQL:       strings.TrimSpace(f.Text),
			})
		}

		inputBuf := strings.TrimSpace(rp.InputBuf)
		activeSQL := inputBuf
		if len(frames) > 0 && frames[0].SQL != "" {
			activeSQL = frames[0].SQL
		}

		curDBName := strings.TrimSpace(rp.CurrentDBName)
		if curDBName != "" {
			dbSet[curDBName] = struct{}{}
		}

		processes = append(processes, Process{
			ID:              rp.ID,
			IsVictim:        isVictim,
			SPID:            fallback(rp.SPID, "?"),
			ECID:            fallback(rp.ECID, "0"),
			Status:          rp.Status,
			WaitTimeMs:      waitMs,
			WaitResource:    rp.WaitResource,
			LockMode:        rp.LockMode,
			LogUsed:         logUsed,
			Priority:        fallback(rp.Priority, "0"),
			TranCount:       fallback(rp.TranCount, "0"),
			TransactionName: rp.TransactionName,
			IsolationLevel:  fallback(rp.IsolationLevel, "unknown"),
			ClientApp:       fallback(strings.TrimSpace(rp.ClientApp), "-"),
			HostName:        fallback(strings.TrimSpace(rp.HostName), "-"),
			LoginName:       fallback(strings.TrimSpace(rp.LoginName), "-"),
			CurrentDB:       rp.CurrentDB,
			CurrentDBName:   curDBName,
			Frames:          frames,
			InputBuf:        inputBuf,
			ActiveSQL:       activeSQL,
		})
	}

	if victimSPID == "" && len(processes) > 0 {
		victimSPID = processes[0].SPID
	}

	resources := make([]Resource, 0, len(rd.ResourceList.Items))
	for i, ri := range rd.ResourceList.Items {
		objName := ri.Attrs["objectname"]
		if objName == "" {
			objName = ri.Attrs["nodeId"]
		}
		if objName == "" {
			objName = ri.Type
		}
		var dbFromObj, shortObj string
		if parts := strings.Split(objName, "."); len(parts) >= 3 {
			dbFromObj = parts[0]
			shortObj = strings.Join(parts[1:], ".")
		} else {
			shortObj = objName
		}
		if dbFromObj != "" {
			dbSet[dbFromObj] = struct{}{}
		}

		hobt := ri.Attrs["hobtid"]
		if hobt == "" {
			hobt = ri.Attrs["associatedObjectId"]
		}
		resID := ri.Attrs["id"]
		if resID == "" {
			resID = fmt.Sprintf("res-%d", i)
		}

		owners := ri.Owners
		if owners == nil {
			owners = []ResourceParticipant{}
		}
		waiters := ri.Waiters
		if waiters == nil {
			waiters = []ResourceParticipant{}
		}

		resources = append(resources, Resource{
			ID:          resID,
			Type:        ri.Type,
			DBID:        ri.Attrs["dbid"],
			Database:    dbFromObj,
			ObjectName:  objName,
			ShortObject: shortObj,
			IndexName:   ri.Attrs["indexname"],
			HoBTID:      hobt,
			FileID:      ri.Attrs["fileid"],
			PageID:      ri.Attrs["pageid"],
			Slot:        ri.Attrs["slot"],
			Mode:        ri.Attrs["mode"],
			Owners:      owners,
			Waiters:     waiters,
		})
	}

	databases := make([]string, 0, len(dbSet))
	for db := range dbSet {
		databases = append(databases, db)
	}
	slices.Sort(databases)

	diag := classifyAndDiagnose(processes, resources)
	fp := computeFingerprint(diag.Pattern, resources)

	return &ParsedDeadlock{
		XMLHash:     HashXML(cleanXML),
		Fingerprint: fp,
		VictimSPID:  victimSPID,
		Databases:   databases,
		MaxWaitMs:   maxWaitMs,
		Processes:   processes,
		Resources:   resources,
		Diagnosis:   diag,
	}, nil
}

func computeFingerprint(pattern string, resources []Resource) string {
	parts := make([]string, 0, len(resources))
	for _, r := range resources {
		parts = append(parts, fmt.Sprintf("%s:%s:%s", strings.ToLower(r.Type), strings.ToLower(r.ObjectName), strings.ToLower(r.IndexName)))
	}
	slices.Sort(parts)
	raw := pattern + "::" + strings.Join(parts, "|")
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:16])
}

func classifyAndDiagnose(processes []Process, resources []Resource) Diagnosis {
	var victim *Process
	others := make([]Process, 0, len(processes))
	byID := make(map[string]Process, len(processes))

	for i := range processes {
		p := processes[i]
		byID[p.ID] = p
		if p.IsVictim && victim == nil {
			victim = &processes[i]
		} else {
			others = append(others, p)
		}
	}
	if victim == nil && len(processes) > 0 {
		victim = &processes[0]
		if len(processes) > 1 {
			others = processes[1:]
		} else {
			others = nil
		}
	}

	hasExchange := false
	hasConvert := false
	hasPageLock := false
	hasSOwnerXWaiter := false
	hasXOwnerSWaiter := false
	objSet := make(map[string]struct{})
	idxSet := make(map[string]struct{})

	for _, r := range resources {
		lt := strings.ToLower(r.Type)
		if strings.Contains(lt, "exchange") {
			hasExchange = true
		}
		if lt == "pagelock" {
			hasPageLock = true
		}
		if r.ObjectName != "" {
			objSet[r.ObjectName] = struct{}{}
		}
		if r.IndexName != "" {
			idxSet[r.IndexName] = struct{}{}
		}

		ownerIDs := make(map[string]struct{}, len(r.Owners))
		hasSOwner := false
		hasXOwner := false
		for _, o := range r.Owners {
			ownerIDs[o.ProcessID] = struct{}{}
			if o.Mode == "S" {
				hasSOwner = true
			}
			if o.Mode == "X" || o.Mode == "U" {
				hasXOwner = true
			}
		}
		for _, w := range r.Waiters {
			if strings.EqualFold(w.RequestType, "convert") {
				hasConvert = true
			}
			if _, alsoOwner := ownerIDs[w.ProcessID]; alsoOwner {
				hasConvert = true
			}
			if hasSOwner && (w.Mode == "X" || w.Mode == "U") {
				hasSOwnerXWaiter = true
			}
			if hasXOwner && w.Mode == "S" {
				hasXOwnerSWaiter = true
			}
		}
	}

	distinctObjects := sortedKeys(objSet)
	distinctIndexes := sortedKeys(idxSet)
	spids := make([]string, 0, len(processes))
	apps := make([]string, 0, len(processes))
	for _, p := range processes {
		spids = append(spids, p.SPID)
		if p.ClientApp != "" && p.ClientApp != "-" && !slices.Contains(apps, p.ClientApp) {
			apps = append(apps, p.ClientApp)
		}
	}

	diag := Diagnosis{
		Pattern:      PatternGenericCycle,
		PatternLabel: "Lock Cycle Deadlock",
		Severity:     "error",
	}

	switch {
	case hasExchange:
		diag.Pattern = PatternParallelExchange
		diag.PatternLabel = "Parallelism Exchange"
		diag.Severity = "info"
		diag.Summary = "An intra-query parallelism deadlock occurred across exchange operators while coordinating parallel worker threads."
		diag.Recommendations = []string{
			"Check missing indexes or stale statistics that forced a large parallel scan/hash plan.",
			"Review Cost Threshold for Parallelism and MAXDOP (or test OPTION (MAXDOP 1) on the affected statement).",
		}

	case hasConvert:
		diag.Pattern = PatternLockUpgrade
		diag.PatternLabel = "Lock Conversion (S -> X)"
		diag.Severity = "purple"
		resName := "the same resource"
		if len(resources) > 0 && resources[0].ShortObject != "" {
			resName = resources[0].ShortObject
		}
		iso := "elevated isolation"
		if victim != nil && victim.IsolationLevel != "" {
			iso = victim.IsolationLevel
		}
		diag.Summary = fmt.Sprintf(
			"Multiple sessions (SPID %s) simultaneously held Shared (S) locks on %s under %s and attempted to convert them to Exclusive (X) locks.",
			strings.Join(spids, ", "), resName, iso,
		)
		diag.Recommendations = []string{
			"Acquire an Update lock on the initial read using WITH (UPDLOCK) before updating the row.",
			"Evaluate whether REPEATABLE READ / SERIALIZABLE is strictly necessary, or switch to READ_COMMITTED_SNAPSHOT (RCSI).",
			"Combine read-then-write logic into a single atomic UPDATE ... OUTPUT statement where possible.",
		}

	case len(distinctObjects) == 1 && len(distinctIndexes) >= 2 && hasSOwnerXWaiter && hasXOwnerSWaiter:
		diag.Pattern = PatternBookmarkLookup
		diag.PatternLabel = "Key / Bookmark Lookup"
		diag.Severity = "warning"
		ncIdx := distinctIndexes[1]
		pkIdx := distinctIndexes[0]
		for _, idx := range distinctIndexes {
			if strings.HasPrefix(strings.ToUpper(idx), "PK_") {
				pkIdx = idx
			} else {
				ncIdx = idx
			}
		}
		diag.Summary = fmt.Sprintf(
			"A reader held a Shared (S) lock on nonclustered index %s while performing a Key Lookup to %s, colliding with a concurrent writer updating %s and maintaining %s.",
			ncIdx, pkIdx, pkIdx, ncIdx,
		)
		diag.Recommendations = []string{
			fmt.Sprintf("Create a covering index by adding selected columns to INCLUDE on %s to eliminate the Key Lookup.", ncIdx),
			"Enable READ_COMMITTED_SNAPSHOT (RCSI) on the database so SELECT queries use row versioning instead of acquiring Shared (S) locks.",
		}

	case len(distinctObjects) >= 2 || hasPageLock:
		diag.Pattern = PatternReverseOrder
		if hasPageLock {
			diag.PatternLabel = "Reverse-Order Page Lock"
		} else {
			diag.PatternLabel = "Reverse-Order Lock Cycle"
		}
		diag.Severity = "error"
		shortObjs := make([]string, 0, len(distinctObjects))
		for _, o := range distinctObjects {
			if parts := strings.Split(o, "."); len(parts) >= 2 {
				shortObjs = append(shortObjs, strings.Join(parts[1:], "."))
			} else {
				shortObjs = append(shortObjs, o)
			}
		}
		objText := strings.Join(shortObjs, " and ")
		victimReason := ""
		if victim != nil {
			if len(others) > 0 && victim.LogUsed < others[0].LogUsed {
				victimReason = fmt.Sprintf(" SQL Server selected SPID %s as the victim because it had lower transaction log usage (%d B vs %d B).", victim.SPID, victim.LogUsed, others[0].LogUsed)
			} else {
				victimReason = fmt.Sprintf(" SQL Server rolled back SPID %s with Error 1205.", victim.SPID)
			}
		}
		if hasPageLock {
			diag.Summary = fmt.Sprintf("SPID %s acquired incompatible PAGE locks on %s in opposite order (often caused by a broad scan or missing index).%s", strings.Join(spids, " and "), objText, victimReason)
			diag.Recommendations = []string{
				"Add a selective index for the DELETE / UPDATE predicates so SQL Server takes narrow KEY locks instead of locking entire 8 KB pages.",
				"Batch large DELETE or archive operations into smaller chunks (e.g., TOP 500 per transaction) to shrink lock duration.",
			}
		} else {
			diag.Summary = fmt.Sprintf("SPID %s acquired incompatible write locks on %s in opposite order inside multi-statement transactions.%s", strings.Join(spids, " and "), objText, victimReason)
			appList := "participating workloads"
			if len(apps) > 0 {
				appList = strings.Join(apps, " and ")
			}
			diag.Recommendations = []string{
				fmt.Sprintf("Standardize lock acquisition order across %s so every transaction updates %s in the exact same order.", appList, objText),
				"Keep transactions as short as possible and avoid application work while holding X/U locks.",
				"Inspect the Execution Plans of both queries in Plan Analysis to ensure index seeks rather than broad scans.",
			}
		}

	default:
		diag.Summary = fmt.Sprintf("Circular lock wait detected between SPID %s on %s.", strings.Join(spids, " and "), fallback(strings.Join(distinctObjects, ", "), "contested resources"))
		diag.Recommendations = []string{
			"Inspect lock modes and ensure consistent access ordering across concurrent transactions.",
			"Consider enabling READ_COMMITTED_SNAPSHOT (RCSI) if read/write contention is involved.",
		}
	}

	steps := make([]CycleStep, 0, len(resources)*2+1)
	for _, r := range resources {
		resLabel := r.ShortObject
		if r.IndexName != "" {
			resLabel = fmt.Sprintf("%s [%s]", r.ShortObject, r.IndexName)
		} else if r.PageID != "" {
			resLabel = fmt.Sprintf("%s (Page %s:%s)", r.ShortObject, r.FileID, r.PageID)
		}
		for _, o := range r.Owners {
			p := byID[o.ProcessID]
			steps = append(steps, CycleStep{
				Kind: "hold",
				Text: fmt.Sprintf("SPID %s (%s) holds %s lock on %s", fallback(p.SPID, o.ProcessID), fallback(p.ClientApp, "-"), o.Mode, resLabel),
			})
		}
		for _, w := range r.Waiters {
			p := byID[w.ProcessID]
			verb := "waits for"
			if strings.EqualFold(w.RequestType, "convert") {
				verb = "attempts to convert lock to"
			}
			steps = append(steps, CycleStep{
				Kind: "wait",
				Text: fmt.Sprintf("SPID %s (%s) %s %s lock on %s (%d ms)", fallback(p.SPID, w.ProcessID), fallback(p.ClientApp, "-"), verb, w.Mode, resLabel, p.WaitTimeMs),
			})
		}
	}
	if victim != nil {
		steps = append(steps, CycleStep{
			Kind: "victim",
			Text: fmt.Sprintf("Deadlock monitor breaks cycle -> SPID %s chosen as VICTIM and rolled back (Error 1205, log used: %d B)", victim.SPID, victim.LogUsed),
		})
	}
	diag.Steps = steps

	return diag
}

func sortedKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

func fallback(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}
