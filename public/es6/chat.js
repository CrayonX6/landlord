bsw.configure({
    data: {
        form: null,
        ws: null,
        nickname: null,
        token: null,
        userId: null,
        online: 0,
        chat: null,
        chatHistory: [],
    },
    method: {
        inputNickname() {
            let nickname = prompt("请输入你的昵称")
            if (nickname == null || nickname.length === 0) {
                return this.inputNickname()
            }
            return nickname
        },
        doChat(e) {
            e.preventDefault()
            let that = this
            if (that.chat == null || that.chat.length === 0) {
                return
            }
            that.ws.send(JSON.stringify({
                Authorization: that.token,
                Behavior: "chat",
                Arguments: {
                    message: that.chat
                },
                Time: Math.floor((new Date()).getTime() / 1000)
            }))
            that.chat = null
        }
    },
    logic: {
        initForm(v) {
            v.form = v.$form.createForm(v)
        },
        ws(v) {
            v.nickname = v.inputNickname()
            v.ws = new ReconnectingWebSocket(`ws://${window.location.host}/ws/chat`)
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
                console.log("收到消息", response)
                if (response.Behavior === "register") {
                    [v.token, v.userId] = [response.Arguments.token, response.Arguments.user_id]
                } else if (response.Behavior === "online") {
                    v.chatHistory.push({
                        pOnlineClass: true,
                        bText: `${response.Arguments.nickname} 上线了`,
                    })
                } else if (response.Behavior === "offline") {
                    v.chatHistory.push({
                        pOfflineClass: true,
                        bText: `${response.Arguments.nickname} 下线了`,
                    })
                } else if (response.Behavior === "chat") {
                    v.chatHistory.push({
                        bText: response.Arguments.time,
                        pText: `${response.Arguments.from}: ${response.Arguments.message}`,
                    })
                }
                if (typeof response.Arguments.online !== 'undefined') {
                    v.online = response.Arguments.online
                }

                let chatDiv = $("div#chat")
                chatDiv.animate({scrollTop: chatDiv.prop("scrollHeight")}, 200)
            }
        }
    }
})