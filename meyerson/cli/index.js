class Action {
  static Add = "add";
  static Start = "start";
  static Finished = "finished";
  static Interrupt = "interrupt";
  static Kill = "kill";
  static Del = "del";
  static InterruptRestart = "interrupt-restart";
  static KillRestart = "kill-restart";
  static Refresh = "refresh";
  static Env = "env";
  static Error = "error";
};
class Status {
  static NotStarted = "NOT STARTED";
  static Running = "RUNNING";
  static Finished = "FINISHED";
};

function newProc(name) {
  return {
    "name": "Test Sleep",
    "program": "sleep",
    "dir": "",
    "args": ["5"],
    "env": [],
    "outFilename": "",
    "errFilename": "",
  };
}

function newMsg(action, content) {
  return {"action": action, "content": content};
}
function newMsgProc(action, proc) {
  return {"action": action, "processes": [proc]};
}

const App = {
  data() {
    const url = new URL("/ws", document.location.origin);
    url.protocol = "ws";
    const ws = new WebSocket(url.toString());
    ws.onopen = () => {
      this.refreshProcs();
      this.getGlobalEnv();
    };
    ws.onmessage = this.msgHandler;
    ws.onerror = this.errHandler;
    ws.onclose = (ev) => {
      console.log("Closed:", ev);
      alert("Connection closed");
    };
    return {
      proc: newProc(""),
      editing: false,
      procs: [],

      globalEnv: [],
      showingGlobalEnv: false,

      detailsShowing: false,

      ws: ws
    };
  },

  methods: {
    startNewProc() {
      this.editing = true;
      this.proc = newProc();
    },
    addProc() {
      this.sendMsg(newMsgProc(Action.Add, this.proc));
      this.clearProc();
      this.editing = false;
    },
    delProc(num) {
      this.sendMsg(newMsg(Action.Del, num));
    },
    clearProc() {
      this.proc = newProc();
    },
    getGlobalEnv() {
      this.sendMsg(newMsg(Action.Env));
    },
    refreshProcs() {
      this.sendMsg(newMsg(Action.Refresh));
    },
    findProcOrRefresh(num) {
      const proc = this.procs.find((p) => p.num == num);
      if (proc) {
        return proc;
      }
      this.sendMsg(newMsg(Action.Refresh, num));
    },
    msgHandler(ev) {
      let msg;
      try {
        msg = JSON.parse(ev.data);
      } catch {
        console.log("received bad message:", ev.data);
        return;
      };
      switch (msg.action) {
        case Action.Add:
          for (var proc of msg.processes) {
            this.procs.push(proc);
          }
          break;
        case Action.Start:
          var proc = this.findProcOrRefresh(msg.content);
          if (proc) {
            proc.status = Status.Finished;
          }
          break;
        case Action.Finished:
          var proc = this.findProcOrRefresh(msg.content);
          if (proc) {
            proc.status = Status.Finished;
          }
          break;
        case Action.Interrupt:
          var proc = this.findProcOrRefresh(msg.content);
          if (proc) {
            // TODO
          }
          break;
        case Action.Kill:
          var proc = this.findProcOrRefresh(msg.content);
          if (proc) {
            // TODO
          }
          break;
        case Action.Del:
          const i = this.procs.findIndex((p) => p.num == msg.content);
          if (i != -1) {
            this.procs.splice(i, 1)
          }
          break;
        case Action.Refresh:
          if (msg.content instanceof Array) {
            if (msg.content[0] == -1) {
              this.procs = msg.processes ?? [];
              return;
            }
            for (var num of msg.content) {
              const i = this.procs.findIndexOf((p) => p.num == num);
              if (i != -1) {
                this.procs.splice(i, 1);
              }
            }
          }
          for (var proc of msg.processes) {
            const i = this.procs.findIndexOf((p) => p.num == proc.num);
            if (i != -1) {
              this.procs[i] = proc;
            } else {
              this.procs.push(proc);
            }
          }
          this.sortProcs();
          break;
        case Action.Env:
          this.globalEnv = msg.content;
          break;
        case Action.Error:
          alert(`Error received: ${msg.error}`);
          break;
        default:
          console.log("received unexpected message:", msg);
      }
    },
    errHandler(ev) {
      alert("Websocket Error");
      console.log(`ERROR: ${ev}`);
    },
    sendMsg(msg) {
      this.ws.send(JSON.stringify(msg));
    },
    collapseExpandDetails() {
      for (const details of document.querySelectorAll(".proc-details")) {
        if (this.detailsShowing) {
          details.removeAttribute("open");
        } else {
          details.setAttribute("open", true);
        }
      }
      this.detailsShowing = !this.detailsShowing;
    },
    sortProcs() {
      this.procs.sort((a, b) => {
        if (a.num > b.num) {
          return 1;
        } else if (a.num < b.num) {
          return -1;
        }
        return 0;
      });
    }
  }
};
Vue.createApp(App).mount("#app");