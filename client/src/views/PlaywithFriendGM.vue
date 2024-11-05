<template>
    <v-container class="pb-0">
        <!-- Loader will load till it makes a ws connection scope of this undermind as when we mount now we ask for username and we check for socket connection there-->
        <v-row v-if="loader==true" class="text-center" justify="center" align="center">
            <v-progress-circular
                :size="70"
                :width="7"
                color="purple"
                indeterminate
            ></v-progress-circular>
        </v-row>
        <!-- Same as above loader will load -->
        <v-row v-if="loader==true" class="text-center" justify="center" align="center">
            <p>Initializing a room</p>
        </v-row>

        <!-- This v-row is enabled when ws connection is made(technically checks in ws send in method not before) and has to fill up his username if not then "Guest" name is taken -->
        <v-row v-if="loader==false && nameformdone==false" class="text-center d-flex flex-column" align="start">
            <v-container>
                <v-card
                    class="mx-auto"
                    max-width="344"
                    variant="outlined"
                >
                    <v-card-item>
                        <div>
                            <div class="mb-1">
                                <h3>Care to enter your name</h3>
                                <small>This is only for your friend to recognize you <br> but you can skip it</small>
                                <br><br>
                                <v-form ref="form" @submit.prevent="sendNewName">
                                    <v-text-field
                                        v-model="firstName"
                                        label="Good Name "
                                        :rules="firstNameRules"
                                        required
                                    ></v-text-field>
                                    <v-btn type="submit" block class="mt-2" >Submit 👍</v-btn>
                                </v-form>
                                <v-btn @click="notSendNewName" block class="mt-2">No need of name 😁</v-btn>
                            </div>
                        </div>
                        </v-card-item>
                    </v-card>
                </v-container>
        </v-row>

        <!-- This v-row starts only when user is done with processing his name and websocket connection is established (loader)-->
        <v-row v-if="loader==false && nameformdone==true && isthegamestarted == false && roomname!=null" class="text-center d-flex flex-column" align="start">
            <!-- Generate users avatar -->
            <v-col v-if="users.length>0 &&  gotthrownout != true">
                <v-container>
                    <v-row align="start" style="height: auto;" class="">
                    <v-col align-self="center" v-for="user in users" :key="user" class="px-0 mx-0">
                        
                            <v-avatar  :color="user.color" class="px-0">
                                {{ user.username[0] }}{{ user.username[1] }}
                                <v-tooltip
                                activator="parent"
                                location="bottom"
                            >
                                {{ user.username }}@<small>{{ user.userslug }}</small>
                            </v-tooltip>
                            </v-avatar>
                            
                    </v-col>
                </v-row>
                </v-container>
            </v-col>
            
            <!-- User who failed to enter room will be notify too -->
            <v-col class="mb-4"  v-for="alert in failedalerts" :key="alert">
                <v-alert
                    type="error"
                    closable
                >{{ alert.message_body.message }}</v-alert>
            </v-col>
            <v-col class="mb-4"  v-for="alert in notifyalert" :key="alert">
                <v-alert
                    type="info"
                    closable
                >{{ alert.message_body.message }}</v-alert>
            </v-col>

            <!-- This column contains containers room occupied(start game)/unoccupied(share the url)/fail to enter(occupied so no user entry) -->
            <v-col class="mb-4">
                
                <!-- The room is not yet occupied so url sharing option is available here -->
                <v-container v-if="playercount < 10 && gotthrownout != true && israndomgame != true">
                    <h4  class="display-2 font-weight-bold mb-3">
                        Share this url to your friend till we wait in lobby 😃 
                    </h4>
                    <h4  class="display-2 font-weight-bold mb-3">Max 10 players allowed</h4>
                    <v-card
                        class="mx-auto"
                        max-width="344"
                        variant="outlined"
                    >
                        <v-row justify="end" class="pt-1">
        
                            <v-col cols="2">
                                <v-btn @click="copyText"  density="compact" icon="mdi-content-copy" size="small"></v-btn>
                            </v-col>
                        </v-row>
                        
                        <v-card-item>
                        <div>

                            <div class="text-h6 mb-1">
                                    {{url + $route.fullPath}}
                            </div>
                           
                        </div>
                        </v-card-item>

                    </v-card>
                </v-container>
                <v-snackbar
                    v-model="copytoclipboard"
                    :timeout="copytoclipboardtimeout"
                    >
                    {{ copytoclipboardmessage }}
                </v-snackbar>

                <!-- Start the game button will come up once the room is occupied  -->
                <v-container v-if="playercount >= 2 && israndomgame != true && gotthrownout != true">
                    <v-card
                        class="mx-auto"
                        max-width="344"
                        variant="outlined"
                    >
                        <v-card-item>
                        <div>
                            <div class="text-h6 mb-1">
                                <h3>Hold tight we gotta go in lads 🐱‍🏍</h3>
                                <small v-if="isroomaker && startthegamebuttonloader == false">your friend is waiting let's start this game 😃</small>
                                <!-- Roommaker can start the game -->
                                <div v-if="isroomaker==true && gamesettingsapplied==false">
                                    <v-form ref="form" @submit.prevent="applyRoomSettings">
                                    <!-- Player Limit Range -->
                                    <v-row>
                                    <v-col cols="12">
                                        <v-slider
                                        v-model="playerlimit"
                                        :min="2"
                                        :max="10"
                                        label="Player Limit"
                                        thumb-label="always"
                                        step="1"
                                        class="mt-3"
                                        >
                                        <template v-slot:append>
                                            <span>{{ playerlimit }} Players</span>
                                        </template>
                                        </v-slider>
                                    </v-col>
                                    </v-row>

                                    <!-- Time Range -->
                                    <v-row>
                                    <v-col cols="12">
                                        <v-slider
                                        v-model="gameduration"
                                        :min="20"
                                        :max="180"
                                        color="orange"
                                        label="Game Time (seconds)"
                                        thumb-label="always"
                                        step="10"
                                        class="mt-3"
                                        >
                                        <template v-slot:append>
                                            <span>{{ gameduration }} sec</span>
                                        </template>
                                        </v-slider>
                                    </v-col>
                                    </v-row>

                                    
                                    <v-btn
                                    v-if="isroomaker==true"
                                    block
                                    class="mt-3"
                                    color="purple"
                                    rounded
                                    @click="applyRoomSettings"
                                    >
                                    Apply Room Settings
                                    </v-btn>
                                    </v-form>

                                </div>
                                <v-btn v-if="isroomaker==true && startthegamebuttonloader == false" block class="mt-3" color="purple" rounded @click="startTheGame">Start the Game 🥁</v-btn>
                                <v-btn v-if="isroomaker==true && startthegamebuttonloader == true" block class="mt-3" color="purple" rounded outlined>
                                <v-progress-circular
                                        indeterminate
                                        color="white"
                                ></v-progress-circular>&nbsp;
                                We are into it
                                </v-btn>
                                <small v-if="isroomaker==false" block class="mt-2" >your friend will start the game so hold tight</small>
                            </div> 
                        </div>
                        </v-card-item>
                    </v-card>
                </v-container>

                <!-- A user attempting to enter occupied room (max entry user limit) is thrown out -->
                <v-container v-if="gotthrownout == true && israndomgame != true">
                    <v-card
                        class="mx-auto"
                        max-width="344"
                        variant="outlined"
                    >
                        <v-card-item>
                        <div>
                            <div class=" mb-1">
                                <h3>This room is occupied 🚧 </h3>
                                <br><hr><br>
                                <h4>This might be the reasons 😕 </h4>
                                <br>
                                <small>
                                    <v-list align="left">
                                    <v-list-item>
                                        😥 You got late and game is started or it got ended
                                    </v-list-item>
                                    <v-list-item>
                                        🐱‍👤 Your room got hijacked 
                                    </v-list-item>
                                    <v-list-item>
                                        👩‍💻 Something in server blew up and causing this issue
                                    </v-list-item>
                                    </v-list>
                                </small>
                                <br>
                                <h5>Just try to create new room cause we don't got any support team to help you out 😛</h5>
                            </div>
                        </div>
                        </v-card-item>
                    </v-card>
                </v-container>

                <!-- The random room is not yet occupied so wait in lobby -->
                <v-container v-if="playercount < 10 && gotthrownout != true && israndomgame == true">
                    <h4  class="display-2 font-weight-bold mb-3">
                        We wait in lobby till we get a new player in 😃 
                    </h4>
                    <v-card
                        class="mx-auto"
                        max-width="344"
                        variant="outlined"
                    >
                        <v-card-item>
                            <h5>2 player random game</h5>
                        </v-card-item>

                    </v-card>
                </v-container>
            </v-col>
            
            
        </v-row>

        <!-- Game container when loader is false nameformdone is true isthegamestarted is true isthgamestoped is false -->
        <v-row v-if="loader==false && nameformdone==true && isthegamestarted == true && isthegamestoped == false && roomname!=null && gotthrownout != true" style="height: auto;" class="text-center d-flex flex-column pb-0" align="start" >
            <v-container max-width="500" class="mt-0 pt-0 pb-0 mb-0" style="height:auto;">
                <v-row align="start" style="height: auto;" class="">
                    <v-col align-self="center">
                        <v-card
                            class="mx-auto pb-0 mb-0"
                            variant="text"
                        >
                            <div class="d-flex flex-row  justify-center">
                                <v-chip
                                    pill
                                    link
                                    v-for="user in users" :key="user"
                                >
                                    {{ user.username[0] }}{{ user.username[1] }}
                                    <v-avatar  :color="user.color" end>
                                        {{ user.score }}
                                    </v-avatar>
                                </v-chip>
                            </div>
                        </v-card>
                    </v-col>
                </v-row>
                <v-row align="center" style="height: auto;" class="mt-0 pt-0">
                    <v-col align-self="center">
                        <v-progress-linear
                                v-model="gameloaderticker"
                                :color="gameloadercolor"
                            ></v-progress-linear>
                        <v-card
                            class="mx-auto"
                            variant="outlined"
                        >
                            
                            <v-row align="end" style="height: 65vh;  overflow:auto;" ref="chatscroll">
                                <v-col>

                                    <!-- Event Bot -->
                                    <div v-for="(message, index) in gamemessages" :key="index"  class="d-flex flex-row justify-end pa-2">
                                        <v-card v-if="message.user=='you'" class="ml-auto rounded-bs-xl mr-3"
                                            max-width="344"
                                            :color="message.color"
                                            variant="elevated">
                                            <v-card-item>
                                                <span class="blue--text mr-3" v-html="message.message"></span>
                                            </v-card-item>
                                        </v-card>
                                        <v-avatar :color="message.color" size="36">
                                            {{ message.useravatar }}
                                        </v-avatar>
                                        <v-card v-if="message.user!='you'" class="mr-auto rounded-be-xl ml-3"
                                            max-width="344"
                                            :color="message.color"
                                            variant="elevated">
                                            <v-card-item>
                                                <span class="blue--text mr-3" v-html="message.message"></span>
                                            </v-card-item>
                                        </v-card>
                                        
                                        
                                    </div> 
                                </v-col>
                            </v-row>
                        </v-card>
                    </v-col>
                </v-row>
                <v-row align="end" style="height: auto;" class="my-0 py-0">
                    <v-col align-self="center" class="mr-0 pr-0">
                        <v-card
                            class="mx-auto pb-0 mb-0"
                            variant="text"
                        >
                            <div class="d-flex flex-row align-center">
                                <v-text-field
                                    label="Enter your word"
                                    style="min-height: auto;"
                                    class="pt-1 pb-0 mb-0 pr-2"
                                    variant="solo"
                                    density="compact"
                                    hide-details
                                    v-model="ingamemessage"
                                    :disabled="usercanentermessage ? false:true"
                                    @keydown.enter.prevent="sendinGameMessage()"
                                    >
                                </v-text-field>
                                <v-btn :disabled="usercanentermessage ? false:true" class="px-2 mr-2" rounded variant="elevated" color="blue" @click="sendinGameMessage()"><v-icon icon="mdi-send"></v-icon></v-btn>
                                <!-- ticker -->
                                <v-avatar color="red" size="small" v-if="usercanentermessage==true">{{ usercanentermessagetimer }}</v-avatar>
                            </div>
                        </v-card>
                    </v-col>
                </v-row>
            </v-container>
        </v-row>

        <!-- Game ends with score board and leave room button -->
        <v-row v-if="loader==false && nameformdone==true && isthegamestarted == true && isthegamestoped == true && gotthrownout != true" style="height: auto;" class="text-center d-flex flex-column pb-0" align="start" >
            <v-container>
                <v-card
                    class="mx-auto"
                    max-width="400"
                    max-height="600"
                    variant="outlined"
                >
                    <v-card-item>
                        <div>
                            <div class="mb-1">
                                <h2>Game board 🏆</h2>
                            </div>
                        </div>
                    </v-card-item>
                    <v-divider
                          :thickness="0.5"
                          class="border-opacity-25 mx-2"
                    ></v-divider>
                    <v-card-item>
                        <v-table
                            fixed-header
                        >
                            <thead>
                            <tr>
                                <th class="text-left">
                                Player
                                </th>
                                <th class="text-left">
                                Score
                                </th>
                            </tr>
                            </thead>
                            <tbody>
                            <tr
                                v-for="(user,index) in userstats"
                                :key="user.slug"
                            >
                                <td v-if="index==0" class="text-left">{{ user.username }}<small>@{{user.userslug}}</small> 🥇</td>
                                <td v-if="index==1" class="text-left">{{ user.username }}<small>@{{user.userslug}}</small> 🥈</td>
                                <td v-if="index==2" class="text-left">{{ user.username }}<small>@{{user.userslug}}</small> 🥉</td>
                                <td v-if="index>2" class="text-left">{{ user.username }}<small>@{{user.userslug}}</small></td>
                                <td class="text-left">{{ user.score }}</td>
                            </tr>
                            </tbody>
                        </v-table>
                    </v-card-item>
                    <v-divider
                          :thickness="0.5"
                          class="border-opacity-25 mx-2"
                    ></v-divider>
                    <v-card-item>
                        <div>
                            <h4>Words guessed:-</h4>
                            <div class="mb-1">
                               <v-chip class="ma-2" label v-for="word in wordslist" :key="word">
                                    {{ word }}
                               </v-chip>
                            </div>
                        </div>
                    </v-card-item>
                </v-card>
                <v-card
                    class="mx-auto"
                    max-width="200"
                    variant="text"
                >
                <v-btn  block class="mt-3" color="purple" rounded @click="leaveTheRoom" prepend-icon="mdi-logout">Leave the room</v-btn>
                </v-card>
            </v-container>
        </v-row>
        <!-- <v-row  align="end" class="text-center" justify="center" >
            <p>This is a demo room</p>
        </v-row> -->
    </v-container>
</template>
  
<script>
//import ws from "../websocket"
import router from "../router/index"

export default {
    /* eslint-disable no-useless-escape */
  name:"PlaywithFriendGM",
  data() {
        return {
        copytoclipboard: false,
        copytoclipboardmessage: 'URL copied to clipboard',
        copytoclipboardtimeout: 2000,
        ws: null,
        url: process.env.VUE_APP_BASE_URL,
        roomname: this.$route.query.room,
        israndomgame: false,
        isthegamestarted: false, //game starts close all other container and focus on this
        gamestarttimer: null,
        totalGameTime: null,
        gameloaderticker: 0,
        gameloadercolor: "blue",
        isthegamestoped: false, //when game stops either due to playercount decreased or game ends
        ingamemessage:"", //message send in chat
        gamemessages:[], //messages send in game collected by handler bot ,users
        gamemessage: {
            user:"", //you,bot,or other user
            message:"",
            useravatar:"",
            usercolor:"",
        },
        usercanentermessage:false,//user cannot enter message unless it is set to true
        usercanentermessagetimer:null,//this timer is assigned by bot usercanentermessage is true till that time only

        active:null,
        gotthrownout:null, //hmm might be due to he was ntering when room is occupied
        messages:[],
        startthegamebuttonloader:false,
        isroomaker: false, //only can send message to start the game if this is true
        newletter: "", //if its empty then will get new letter
        userguessedbeforeticker: false, //user guessed before timer ends
        gameticker:null,
        nextclientslug: "", //if its empty then new client
        alerts:[],
        notifyalert:[],
        wordslist:[],
        userstats: [], //[{username:,userslug:,color:}]
        users: [], //[{username:,userslug:,color:}]
        user: {
            username:"",
            userslug:"",
            color:"",
            score:"",

        },
        newusername:"",
        youruserslug:"", //userslug of the client
        failedalerts:[],
        playercount:0,
        playerlimit:3,
        gameduration:30,
        gamesettingsapplied:false,
        loader: false,
        nameformdone: false,
        colors:["blue","yellow","red","orange","purple"],
        alertmessage:null,
        firstName: '',
        firstNameRules: [
            value => {
                let username = value.trim()
                if (username?.length > 3 && username?.length < 10) return true
                return 'name must be between 3 to 10 letters.'
            },
            
        ],
        }
    },
    methods: {
        handleNewMessage(event) {
            let data = event.data;
            data = data.split(/\r?\n/);
            for (let i = 0; i < data.length; i++) {
                let msg = JSON.parse(data[i]);
                // display the message in the correct room.

                if (msg.action=="join-room-notify" && msg.target == this.roomname) {
                    
                    this.alerts.push(msg)
                }
                if (msg.action=="know-yourself") {
                    this.waitForSocketConnection(this.ws, function() {
                    this.ws.send(JSON.stringify({ action: 'set-client-name', message_body:{ setname: this.firstName}, target: this.roomname }));
                }.bind(this));
                    this.youruserslug = msg.message_body.sender
                }
                if (msg.action=="fail-join-room-notify") {
                    this.gotthrownout = true
                    this.failedalerts.push(msg)
                }
                if (msg.action=="found-random-room-notify") {
                    //console.log(msg)
                    this.roomname = msg.message_body.roomname
                }
                if (msg.action=="is-room-maker") {
                    this.isroomaker = true
                }
                if (msg.action=="client-list-notify" && msg.target == this.roomname) {
                    this.users=[]
                    this.playercount = msg.message_body.clientsinroomessage.length

                    msg.message_body.clientsinroomessage.forEach(element => {
                        this.user = {
                            username : element.name,
                            userslug : element.slug,
                            color: element.color,
                            score: element.score
                        }
                        
                        this.users.push(this.user)
                    });
                }

                if (msg.action=="room-bot-greetings" && msg.target == this.roomname) {
                    if (this.gotthrownout==true) {
                        return
                    }
                    this.isthegamestarted = true
                    this.gamestarttimer = this.gameduration
                    this.totalGameTime = this.gameduration
                    this.gameCountDownTimer()
                }
                if (msg.action=="room-bot-end-game" && msg.target == this.roomname) {
                    this.isthegamestoped = true
                    if( msg.message_body.word_list.length>0){
                        this.wordslist = msg.message_body.word_list
                    }
                    this.users=[]
                    msg.message_body.client_list.forEach(element => {
                        this.user = {
                            username : element.name,
                            userslug : element.slug,
                            color: element.color,
                            score: element.score
                        }
                        
                        this.userstats.push(this.user)
                    });

                }
                if (msg.action=="send-message" && msg.target == this.roomname) { //message from room server (user,bot,you)
                    if (msg.sender == "bot-of-the-room") {
                        this.gamemessage = {
                            user : "bot",
                            useravatar: "🤖",
                            message: msg.message_body.message,
                            color:"purple darken-4"
                        }
                        this.gamemessages.push(this.gamemessage)
                    } else if (msg.sender == this.youruserslug) {
                        this.gamemessage = {
                            user : "you",
                            useravatar: msg.message_body.useravatar,
                            message: msg.message_body.message,
                            color:msg.message_body.color
                        }
                        this.gamemessages.push(this.gamemessage)

                    } else {
                        this.gamemessage = {
                            user : "other",
                            useravatar: msg.message_body.useravatar,
                            message: msg.message_body.message,
                            color:msg.message_body.color
                        }
                        this.gamemessages.push(this.gamemessage)

                    }
                    
                }
                if (msg.action=="room-client-message" && msg.target == this.roomname) { //message from client  must be when new word by him is broadcasted to all
                    //message
                    //sender
                }

                if (msg.action=="room-setting-applied" && msg.target == this.roomname) {
                    this.notifyalert.push(msg)
                    if(this.isroomaker) {
                        this.gamesettingsapplied = true
                    }
                }

                if (msg.action=="message-by-bot" && msg.target == this.roomname) { //message from room server 
                    this.gamemessage = {
                            user : "bot",
                            useravatar: "🤖",
                            message: msg.message_body.message,
                            color:"purple darken-4"
                        }
                    this.gamemessages.push(this.gamemessage)
                 
                    for (i =0; i< msg.message_body.clientstats.length; i++) {
                        
                        if(this.users[i].userslug == msg.message_body.clientstats[i].slug) {
                            this.users[i].score = msg.message_body.clientstats[i].score
                        }
                    }
                    if (msg.message_body.letter != "") {
                        this.newletter = msg.message_body.letter
                    }
                    if (msg.message_body.whichclientturn != null) {
                        if (msg.message_body.whichclientturn.slug== this.youruserslug) {
                            if(this.gameticker) {
                                //console.log("user has left over glitch time")
                                clearTimeout(this.gameticker)
                            }
                            this.usercanentermessagetimer = msg.message_body.timer 
                            this.inGamecountDownTimer()

                        } else {
                            this.usercanentermessagetimer = null
                            this.usercanentermessage = false
                        }
                    } else { //then its for all
                        this.usercanentermessagetimer = msg.message_body.timer 
                        this.inGamecountDownTimer()
                    }
                }

                this.$nextTick(() => {
                    if (this.$refs.chatscroll!= undefined) {
                    this.$refs.chatscroll.$el.scrollTop = this.$refs.chatscroll.$el.scrollHeight;
                    }
                });
                
            }
        },
        connectToWebsocket() {
            this.ws = new WebSocket(process.env.VUE_APP_WEBSOCKET_PRO_URL);
            //console.log(process.env);
            this.ws.addEventListener('message', (event) => { this.handleNewMessage(event) });
        },
        connectToGameRoom() {
            if (this.roomname != undefined) {
                //console.log(this.roomname)
                if (this.roomname.length>10) {
                    alert("Roomname length should be equal to 10")
                } else {
                    var format = /[ `!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]/;
                    if (format.test(this.roomname)==false){
                        this.waitForSocketConnection(this.ws, function() {
                            this.ws.send(JSON.stringify({ action: 'join-room', message_body: {roomname: this.roomname,playerlimit:3,setplayername: this.firstName} , target: 'mesh-global' }));
                        }.bind(this));
                    } else {
                        alert("Roomname not valid")
                    }
                }
            } else {
                this.israndomgame = true
                this.waitForSocketConnection(this.ws, function() {
                    this.ws.send(JSON.stringify({ action: 'join-random-room',message_body: {roomname: this.makeroom(10)} , target: 'mesh-global' }));
                }.bind(this));
            }
        },
        makeroom(length) {
            let result = '';
            const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
            const charactersLength = characters.length;
            let counter = 0;
            while (counter < length) {
            result += characters.charAt(Math.floor(Math.random() * charactersLength));
            counter += 1;
            }
            return result;
        },
        startTheGame() {
            this.waitForSocketConnection(this.ws, function() {
                this.ws.send(JSON.stringify({ action: 'start-the-game',  message_body:{ message: this.roomname},target: this.roomname ,sender:this.userslug }));
            }.bind(this));
            this.startthegamebuttonloader = true
        },
        applyRoomSettings() {
            this.waitForSocketConnection(this.ws, function() {
                this.ws.send(JSON.stringify({ action: 'room-settings',  message_body:{ game_duration:  String(this.gameduration),player_limit: String(this.playerlimit)},target: this.roomname ,sender:this.userslug }));
            }.bind(this));
            this.applyRoomSettings = true
        },
        leaveTheRoom() {
            this.waitForSocketConnection(this.ws, function() {
                this.ws.send(JSON.stringify({ action: 'leave-room', message_body:{message:this.roomname },target:this.roomname}));
            }.bind(this));
            router.back();

        },
        
        async sendNewName() {
            const { valid } =  await this.$refs.form.validate()
            if (valid) {
                this.nameformdone = true
                this.connectToGameRoom()
            }
        },
        notSendNewName() {
            this.firstName = "Guest"
            this.nameformdone = true
            this.connectToGameRoom()
        },
        waitForSocketConnection(socket, callback){
            setTimeout(
                function(){
                    this.loader=true
                    if (socket.readyState === 1) {
                        if(callback !== undefined){
                            this.loader = false
                            callback();
                        }
                        return;
                    } else {
                        this.waitForSocketConnection(socket,callback);
                    }
                }.bind(this), 5);
        },
        sendinGameMessage() { //send to websocket client in Go
            if(this.ingamemessage !== "") {
               
                if (this.newletter !="") { //game is on and the bot send a letter
                    this.usercanentermessage = false
                    this.usercanentermessagetimer = null
                    
                    this.userguessedbeforeticker = true
                    // to the bot
                    this.waitForSocketConnection(this.ws, function() {
                        this.ws.send(JSON.stringify(
                            {
                                action:"attempt-word",
                                target:this.roomname,
                                message_body:{message:this.ingamemessage.trim()},
                            })); //send it to websocket
                        this.ingamemessage=""
                    }.bind(this))
                } else {
                    this.waitForSocketConnection(this.ws, function() {
                        this.ws.send(JSON.stringify({action:"send-message",target:this.roomname,message_body:{message:this.ingamemessage}})); //send it to websocket
                        this.ingamemessage=""
                    }.bind(this))
                }
            }
        },
        gameCountDownTimer() {
        if (this.gamestarttimer > 0) {
        setTimeout(() => {
            this.gamestarttimer -= 1;

            // Calculate the percentage of remaining time
            let timerPercentage = (this.gamestarttimer / this.totalGameTime) * 100;

            // Calculate the increment per second dynamically based on total game time
            let incrementPerSecond = 100 / this.totalGameTime;  // e.g., 100 / 180 = 0.555 for 180 seconds

            // Increment the gameloaderticker by the dynamic value
            this.gameloaderticker += incrementPerSecond;

            // Change color based on the percentage
            if (timerPercentage <= 90 && timerPercentage > 30) {
                this.gameloadercolor = "yellow";
            } else if (timerPercentage <= 30) {
                this.gameloadercolor = "red";
            }

            // Recursive call for the countdown
            this.gameCountDownTimer();
        }, 1000);
    }
        },
        inGamecountDownTimer () {
            this.usercanentermessage = true
            if (this.usercanentermessagetimer > 0) {
                this.gameticker = setTimeout(() => {
                    this.usercanentermessagetimer -= 1
                    this.inGamecountDownTimer()
                }, 1000)
            } else {
                this.usercanentermessage = false
                this.usercanentermessagetimer = null
                if (this.newletter == "") { //start the game one time only it should occure
                    //
                } 
                else { //doubt here ending the ticekr in worng way in sendmessage
                    // to the bot
                    this.userguessedbeforeticker = false //if it was true then it was skipped but now we need to recalcuate
                }
            }
        },
        copyText () {
            this.copytoclipboard = true
            let textToCopy = process.env.VUE_APP_BASE_URL + this.$route.fullPath
            navigator.clipboard.writeText(textToCopy);
        }


    },
    mounted: function() {
        this.connectToWebsocket()

    },
   
    beforeRouteLeave(to, from, next) {
        if(this.roomname != null && this.nameformdone == true){
            this.waitForSocketConnection(this.ws, function() {
                this.ws.send(JSON.stringify({ action: 'leave-room', target: this.roomname }));
            }.bind(this));
        }
        this.ws.close()
        next(); 
    },
    beforeRouteUpdate(to, from, next) {
        this.ws.close()
        next();
    }

}
</script>