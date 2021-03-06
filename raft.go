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

import "sync"
import (
	"labrpc"
	"time"
	"math/rand"
	)

// import "bytes"
// import "labgob"



//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in Lab 3 you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh; at that point you can add fields to
// ApplyMsg, but set CommandValid to false for these other uses.
//
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

//
// A Go object implementing a single Raft peer.
//

const (
	Leader = 1
	Candidate = 2
	Follower = 3
	Quited = 4
)

type Log struct {

}

type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	currentTerm		int
	votedFor		int
	log				[]Log

	commitIndex		int
	lastApplied		int

	nextIndex		[]int
	matchIndex		[]int

	role			int

	applyCh			chan ApplyMsg
	quitCh			chan int
	heartBeatsCh		chan int
	leaderElectionCh	chan int
	transformCh			chan int

	leaderElectionTimeout	time.Duration

}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (2A).
	term = rf.currentTerm
	isleader = (rf.role == Leader)
	return term, isleader
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
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//
type AppendEntriesArgs struct {
	Term 	int
	LeaderId	int
	PrevLogIndex	int
	PrevLogTerm		int
	Entries			[]Log
	LeaderCommit	int
}

type AppendEntriesReply struct {
	Term	int
	Success bool
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.heartBeatsCh<-1
}

type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term	int
	CandidateId	int
	LastLogIndex	int
	LastLogTerm		int
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// Your data here (2A).
	Term 	int
	VoteGranted  bool
}

//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()

	rf.checkTerm(args.Term)
	if rf.votedFor == -1 {
		reply.VoteGranted = true
		rf.votedFor = args.CandidateId
	} else if rf.votedFor == rf.me {
		if args.Term > rf.currentTerm {
			reply.VoteGranted = true
			return
		}
	} else {
		reply.VoteGranted = false
	}
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
type voteInfo struct {
	mu          sync.Mutex
	voteRecord	map[int]bool
	cnt         int

}

func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, vote *voteInfo) bool {
	reply := RequestVoteReply{}
	ok := rf.peers[server].Call("Raft.RequestVote", args, &reply)
	if reply.VoteGranted == true {
		if vote.voteRecord[server] == false {
			vote.cnt++
			vote.voteRecord[server] = true
		}
	}
	if vote.cnt >= len(rf.peers) / 2 + 1 {
		if len(rf.transformCh) == 0 {
			rf.transformCh <-Leader
		}
	}
	return ok
}


//c
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).


	return index, term, isLeader
}

//
// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (rf *Raft) Kill() {
	// Your code here, if desired.
	rf.quitCh<-1
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

	// Your initialization code here (2A, 2B, 2C).
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.leaderElectionTimeout = time.Duration(1000 + rand.Int63()%150) * time.Millisecond

	rf.applyCh = applyCh
	rf.heartBeatsCh = make(chan int, 1)
	rf.quitCh = make(chan int, 1)
	rf.leaderElectionCh = make(chan int, 1)
	rf.transformCh = make(chan int, 1)

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	go rf.runRaft()

	return rf
}

func (rf *Raft) runRaft() {
	state := Follower
	for {
		switch state {
		case Follower:
			state = rf.runAsFollower()
		case Candidate:
			state = rf.runAsCandidate()
		case Leader:
			state = rf.runAsLeader()
		case Quited:
			return
		}
	}
}

func (rf *Raft) runAsLeader() int {
	rf.role = Leader
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			//heartbeats
			for i := range rf.peers {
				if i != rf.me {
					go rf.sendAppendEntries(i, &AppendEntriesArgs{}, &AppendEntriesReply{})
				}
			}
		case nextState := <-rf.transformCh:
			return nextState
		case <-rf.quitCh:
			return Quited
		}
	}
}

func (rf *Raft) runAsFollower() int {
	rf.role = Follower
	timer := time.NewTimer(rf.leaderElectionTimeout)
	for {
		select {
		case <-timer.C:
			return Candidate
		case <-rf.heartBeatsCh:
			timer.Reset(rf.leaderElectionTimeout)
		case <-rf.quitCh:
			return Quited
		}
	}
}

func (rf *Raft) runAsCandidate() int {
	rf.role = Candidate
	//leader election
	vote := voteInfo{}
	vote.mu.Lock()
	vote.voteRecord = make(map[int]bool)
	for i := 0; i < len(rf.peers); i++ {
		vote.voteRecord[i] = false
	}
	vote.cnt = 1
	vote.mu.Unlock()

	rf.votedFor = rf.me
	rf.currentTerm = rf.currentTerm + 1

	for i := 0; i < len(rf.peers); i++ {
		if (i != rf.me) {
			go rf.sendRequestVote(i, &RequestVoteArgs{Term: rf.currentTerm, CandidateId: rf.me}, &vote)
		}
	}
	timer := time.NewTimer(rf.leaderElectionTimeout)
	for {
		select {
		case <-timer.C:
			return Candidate
		case nextState := <-rf.transformCh:
			return nextState
		case <-rf.quitCh:
			return Quited
		}
	}
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}



func (rf *Raft) checkTerm(term int) {
	if rf.currentTerm < term {
		rf.currentTerm = term
		rf.votedFor = -1
		if rf.role != Follower {
			rf.role = Follower
			rf.transformCh <- Follower
		}
	}
}
