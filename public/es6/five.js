bsw.configure({
    data: {
        ws: null,
        nickname: null,
        token: null,
        userId: null,
        online: 0,
        roomNumber: 0,
        blackPlayer: {},
        whitePlayer: {},
        myPlayer: {},
        willDown: 0,
        nextColor: "black",
        chessHistory: {},
        isWin: false,
    },
    method: {
        inputNickname() {
            let nickname = prompt("请输入你的昵称")
            if (nickname == null || nickname.length === 0) {
                return this.inputNickname()
            }
            return nickname
        },
        cloneChess() {
            let that = this
            let offset = bsw.offset($(`div.data > div > div.${that.nextColor}`))
            $(".container > .abs").remove()
            $(".container").append(`<div class="chess abs ${that.nextColor}"></div>`)
            let abs = $(".container > div.abs")
            abs.css({left: offset.left, top: offset.top})
            that.willDown = 0
            that.drop(abs)
        },
        drop(element) {
            let that = this
            if (!element.hasClass(that.myPlayer.chess_color)) {
                return
            }
            element.mousedown(function (event) {
                $(document).mousemove(function (e) {
                    let x = e.pageX - event.offsetX
                    let y = e.pageY - event.offsetY
                    element.css({left: x, top: y})
                    let downX = Math.floor((e.pageX - 70) / 50)
                    let downY = Math.floor((e.pageY - 70) / 50)
                    if (e.pageX > 45 && e.pageX < 70) {
                        downX = 0
                    } else if ((e.pageX - 70) % 50 > 25) {
                        downX += 1
                    }
                    if (e.pageY > 45 && e.pageY < 70) {
                        downY = 0
                    } else if ((e.pageY - 70) % 50 > 25) {
                        downY += 1
                    }
                    if (downX < 0 || downX > 14 || downY < 0 || downY > 14) {
                        that.willDown = 0
                    } else {
                        let willDown = (downY * 15) + (downX + 1)
                        if (that.chessHistory[willDown]) { // 已落子
                            willDown = 0
                        }
                        that.willDown = willDown
                    }
                    if (that.isWin) {
                        return
                    }
                    that.ws.send(JSON.stringify({
                        Authorization: that.token,
                        Behavior: "move",
                        Arguments: {
                            left: x,
                            top: y,
                            willDown: that.willDown,
                        }
                    }))
                })
                $(document).mouseup(function () {
                    $(document).off("mousemove")
                    $(document).off("mouseup")
                    if (that.willDown > 0) {
                        if (that.isWin) {
                            return
                        }
                        that.ws.send(JSON.stringify({
                            Authorization: that.token,
                            Behavior: "chessDown",
                            Arguments: {
                                willDown: that.willDown,
                            }
                        }))
                    }
                })
            })
        },
        resetPlayers(players) {
            let that = this
            for (let member of players) {
                if (member.chess_color === "black") {
                    that.blackPlayer = member
                } else {
                    that.whitePlayer = member
                }
                if (member.player_id === that.userId) {
                    that.myPlayer = member
                }
            }
        }
    },
    logic: {
        ws(v) {
            v.nickname = v.inputNickname()
            v.ws = new ReconnectingWebSocket(`ws://${window.location.host}/ws/five`)
            v.ws.onopen = function () {
                console.log("成功连接服务器。")
                v.ws.send(JSON.stringify({
                    Behavior: "register",
                    Arguments: {
                        nickname: v.nickname
                    }
                }))
            }
            v.ws.onclose = function () {
                console.log("收到关闭请求，成功断开连接。")
            }

            v.ws.onmessage = function (msg) {
                let response = eval(`(${msg.data})`)
                let args = response.Arguments
                if (response.Behavior !== "move") {
                    console.log("收到消息", response)
                }
                if (response.Behavior === "register") {
                    [v.token, v.userId] = [args.token, args.user_id]
                } else if (response.Behavior === "online") {
                    bsw.success(args.message)
                    v.nextColor = args.nextColor
                    v.resetPlayers(args.players)
                    if (args.players.length === 2) {
                        v.cloneChess()
                    }
                    v.chessHistory = args.history
                } else if (response.Behavior === "offline") {
                    bsw.warning(args.message)
                } else if (response.Behavior === "move") {
                    let abs = $(".container > div.abs")
                    abs.css({left: args.left, top: args.top})
                    v.willDown = args.willDown
                } else if (response.Behavior === "chessDown") {
                    v.chessHistory = args.history
                    v.nextColor = args.nextColor
                    v.resetPlayers(args.players)
                    v.cloneChess()
                    if (args.win.length > 0) {
                        v.isWin = true
                        if (args.win === v.myPlayer.chess_color) {
                            bsw.confirm("success", "你赢了，继续保持哟~", 0)
                        } else {
                            bsw.confirm("warning", "你输了，要努力了哟~", 0)
                        }
                    }
                }
                if (typeof args.room_number !== 'undefined') {
                    v.roomNumber = args.room_number
                }
                if (typeof args.online !== 'undefined') {
                    v.online = args.online
                }
            }
        }
    }
})