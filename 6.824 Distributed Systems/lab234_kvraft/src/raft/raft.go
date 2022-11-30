package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	//	"bytes"
	"sync"
	"sync/atomic"
	"time"

	//	"6.824/labgob"
	"6.824/labrpc"
	//
	"math/rand"
)

const (
	heartHeatsInterval int = 50
	electionInterval   int = 550
	follower           int = 1
	candidate          int = 2
	leader             int = 3
	dead               int = 4
)

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 2D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
//
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	state int

	// persistent state on all server
	currentTerm int
	votedFor    int
	log         []LogEntry

	//volatile state on all servers
	commitIndex   int
	lastApplied   int
	lastAppliedAt time.Time

	//volatile state on leader
	nextIndex  []int
	matchIndex []int
}

type LogEntry struct {
	Index int
	Term  int
	Cmd   interface{}
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isLeader bool
	// Your code here (2A).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	term = rf.currentTerm
	isLeader = rf.state == leader
	return term, isLeader
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

//
// A service wants to switch to snapshot.  Only do so if Raft hasn't
// have more recent info since it communicate the snapshot on applyCh.
//
func (rf *Raft) CondInstallSnapshot(lastIncludedTerm int, lastIncludedIndex int, snapshot []byte) bool {

	// Your code here (2D).

	return true
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (2D).

}

//
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// Your data here (2A).
	Term        int
	VoteGranted bool
}

//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()

	reply.Term = rf.currentTerm
	reply.VoteGranted = false
	lastLog := rf.lastLogEntry()
	// DPrintf("server %v (term: %v, votedFor:%v, lastLogTerm:%v, logLength:%v) \nreceived (CandidateId:%v, Term:%v, LastLogTerm:%v, LastLogIndex:%v ", rf.me, rf.currentTerm, rf.votedFor, lastLog.term, len(rf.log), args.CandidateId, args.Term, args.LastLogTerm, args.LastLogIndex)
	// DPrintf("server %v(term:%v) voted for %v, candidate %v(term:%v)", rf.me, rf.currentTerm, rf.votedFor, args.CandidateId, args.Term)
	if args.Term < rf.currentTerm {
		return
	}

	if args.Term > rf.currentTerm {
		rf.toFollower(args.Term)
	}

	upToDate := false

	if args.LastLogTerm > lastLog.Term {
		upToDate = true
	}
	if args.LastLogTerm == lastLog.Term && args.LastLogIndex+1 >= len(rf.log) {
		upToDate = true
	}

	if (rf.votedFor == -1 || rf.votedFor == args.CandidateId) && upToDate {
		reply.VoteGranted = true
	}
}

//
//
//
type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	reply.Success = true
	reply.Term = rf.currentTerm
	if args.Term < rf.currentTerm {
		reply.Success = false
		return
	}
	if args.Term > rf.currentTerm {
		rf.toFollower(args.Term)
	}

	if rf.currentTerm == args.PrevLogTerm {
		reply.Success = false
	}

	// DPrintf("server %v received heartbeat from %v(%v) at term %v(time:%v)", rf.me, args.LeaderId, args.Term, rf.currentTerm, rf.lastAppliedAt)

}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	index := -1
	term := rf.currentTerm
	isLeader := rf.state == leader

	// Your code here (2B).
	if isLeader {
		index = rf.nextLogIndex()
		rf.log = append(rf.log, LogEntry{Index: index, Term: term, Cmd: command})
	}

	return index, term, isLeader
}

//
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

// The ticker go routine starts a new election if this peer hasn't received
// heartsbeats recently.
func (rf *Raft) ticker() {
	for rf.killed() == false {

		// Your code here to check if a leader election should
		// be started and to randomize sleeping time using
		// time.Sleep().

		electionTimeout := electionInterval + rand.Int()%333
		startTime := time.Now()
		time.Sleep(time.Duration(electionTimeout) * time.Millisecond)

		rf.mu.Lock()
		if rf.state == dead {
			rf.mu.Unlock()
			return
		}
		if rf.lastAppliedAt.Before(startTime) {
			if rf.state != leader {
				// DPrintf("server %v start election at term %v\n   st:%v\n   la:%v\n   no:%v", rf.me, rf.currentTerm, startTime, rf.lastAppliedAt, time.Now())
				// DPrintf("server %v start election at term %v", rf.me, rf.currentTerm)
				go rf.attemptElection()
			}
		}
		rf.mu.Unlock()
	}
}

func (rf *Raft) attemptElection() {

	rf.mu.Lock()
	rf.toCandidate()
	lastLog := rf.lastLogEntry()
	args := RequestVoteArgs{
		Term:         rf.currentTerm,
		CandidateId:  rf.me,
		LastLogIndex: lastLog.Index,
		LastLogTerm:  lastLog.Term,
	}
	rf.mu.Unlock()

	votes := 1
	done := false
	for p := range rf.peers {

		if p != rf.me {
			go func(server int) {
				reply := RequestVoteReply{}

				ok := rf.sendRequestVote(server, &args, &reply)

				if !ok {
					return
				}
				rf.mu.Lock()
				defer rf.mu.Unlock()
				if reply.Term > rf.currentTerm {
					rf.toFollower(reply.Term)
					return
				}
				// DPrintf("server %v received %v(%v) from %v at term %v(state: %v)", rf.me, reply.VoteGranted, reply.Term, server, rf.currentTerm, rf.state)
				if reply.VoteGranted {
					votes++
					if done || votes <= len(rf.peers)/2 {
						return
					}
					if rf.state != candidate || rf.currentTerm < reply.Term {
						return
					}
					rf.toLeader()
					// DPrintf("server %v became leader at term %v", rf.me, rf.currentTerm)
					for j := range rf.peers {
						if j != rf.me {
							go rf.OperateLeader(j)
						}
					}
					done = true
				}

			}(p)
		}
	}
}

func (rf *Raft) OperateLeader(pid int) {
	for {
		rf.mu.Lock()
		if rf.state != leader {
			rf.mu.Unlock()
			return
		}
		rf.mu.Unlock()
		go rf.sendAppendEntry(pid)
		time.Sleep(time.Duration(heartHeatsInterval) * time.Millisecond)
	}
}

func (rf *Raft) sendAppendEntry(pid int) {
	rf.mu.Lock()
	prevLogIndex := rf.nextIndex[pid] - 1
	prevLogTerm := -1
	if prevLogIndex >= 0 {
		prevLogTerm = rf.log[prevLogIndex].Term
	}
	args := AppendEntriesArgs{
		Term:         rf.currentTerm,
		LeaderId:     rf.me,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      make([]LogEntry, len(rf.log[prevLogIndex+1:])),
		LeaderCommit: rf.commitIndex,
	}
	copy(args.Entries, rf.log[prevLogIndex+1:])
	rf.mu.Unlock()

	reply := AppendEntriesReply{}
	ok := rf.sendAppendEntries(pid, &args, &reply)
	if !ok {
		// DPrintf("server %v send append to %v at term %v: not ok", rf.me, pid, rf.currentTerm)
		return
	}

	rf.mu.Lock()
	defer rf.mu.Unlock()
	if reply.Term > rf.currentTerm {
		rf.toFollower(reply.Term)
		return
	}
}

func (rf *Raft) toCandidate() {
	rf.state = candidate
	rf.votedFor = rf.me
	rf.lastAppliedAt = time.Now()
	rf.currentTerm++
}

func (rf *Raft) toFollower(newTerm int) {
	rf.state = follower
	rf.votedFor = -1
	rf.lastAppliedAt = time.Now()
	rf.currentTerm = newTerm
}

func (rf *Raft) toLeader() {
	rf.state = leader
	rf.lastAppliedAt = time.Now()
	rf.nextIndex = make([]int, len(rf.peers))
	rf.matchIndex = make([]int, len(rf.peers))
	for i := range rf.peers {
		rf.nextIndex[i] = rf.nextLogIndex()
		rf.matchIndex[i] = 0
	}
}

func (rf *Raft) lastLogEntry() LogEntry {
	if len(rf.log) > 0 {
		return rf.log[len(rf.log)-1]
	} else {
		return LogEntry{
			Index: -1,
			Term:  -1,
		}
	}
}

func (rf *Raft) nextLogIndex() int {
	return len(rf.log)
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.nextIndex = make([]int, len(rf.peers))
	rf.matchIndex = make([]int, len(rf.peers))
	for i := range rf.peers {
		rf.nextIndex[i] = rf.nextLogIndex()
		rf.matchIndex[i] = 0
	}

	// Your initialization code here (2A, 2B, 2C).
	rf.toFollower(0)

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.ticker()

	return rf
}