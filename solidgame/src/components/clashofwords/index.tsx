import { createSignal, createEffect, onMount, onCleanup, For, Show } from 'solid-js';
import { useNavigate, useParams, useLocation } from '@solidjs/router';
import { gsap } from 'gsap';
import {
  Avatar,
  Button,
  Card,
  CardHeader,
  CardContent,
  Chip,
  CircularProgress,
  Container,
  Divider,
  LinearProgress,
  List,
  ListItem,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Alert,
} from "@suid/material";
import PlayerDetails from './PlayerDetails';
import PlayerLobby from './PlayerLobby';
import PlayerLeaderboard from './PlayerLeaderboard';
import { deepOrange } from '@suid/material/colors';

interface User {
  username: string;
  userslug: string;
  color: string;
  score: number;
}

interface GameMessage {
  user: string;
  message: string;
  useravatar: string;
  color: string;
}

interface ClientStats {
  name: string;
  slug: string;
  clientgamemetadata: {
    color: string;
    score: number;
  };
}

interface WebSocketMessage {
  action: string;
  target?: string;
  message?: string | any;
  sender?: {
    name: string;
    slug: string;
    clientgamemetadata?: {
      color: string;
      score: number;
    };
  };
  clientlist?: ClientStats[];
  clientstats?: ClientStats[];
  whichclientturn?: {
    slug: string;
  };
  timer?: number;
  letter?: string;
}

interface AlertMessage {
  message: string;
}

const PlayClashofWords = () => {
  const navigate = useNavigate();
  const params = useParams();
  const location = useLocation();

  // Signals
  const [copytoclipboard, setCopytoclipboard] = createSignal(false);
  const [copytoclipboardmessage, setCopytoclipboardmessage] = createSignal('URL copied to clipboard');
  const [copytoclipboardtimeout, setCopytoclipboardtimeout] = createSignal(2000);
  const [ws, setWs] = createSignal<WebSocket | null>(null);
  const [url, setUrl] = createSignal("http://localhost:8081");
  const [roomname, setRoomname] = createSignal<string | null>(new URLSearchParams(location.search).get('room'));
  const [israndomgame, setIsrandomgame] = createSignal(false);
  const [isthegamestarted, setIsthegamestarted] = createSignal(false);
  const [gamestarttimer, setGamestarttimer] = createSignal<number | null>(null);
  const [gameloaderticker, setGameloaderticker] = createSignal(0);
  const [gameloadercolor, setGameloadercolor] = createSignal('blue');
  const [isthegamestoped, setIsthegamestoped] = createSignal(false);
  const [ingamemessage, setIngamemessage] = createSignal('');
  const [gamemessages, setGamemessages] = createSignal<GameMessage[]>([]);
  const [usercanentermessage, setUsercanentermessage] = createSignal(false);
  const [usercanentermessagetimer, setUsercanentermessagetimer] = createSignal<number | null>(null);
  const [gotthrownout, setGotthrownout] = createSignal<boolean | null>(null);
  const [startthegamebuttonloader, setStartthegamebuttonloader] = createSignal(false);
  const [isroomaker, setIsroomaker] = createSignal(false);
  const [newletter, setNewletter] = createSignal('');
  const [userguessedbeforeticker, setUserguessedbeforeticker] = createSignal(false);
  const [gameticker, setGameticker] = createSignal<number | null>(null);
  const [wordslist, setWordslist] = createSignal<string[]>([]);
  const [userstats, setUserstats] = createSignal<User[]>([]);
  const [users, setUsers] = createSignal<User[]>([]);
  const [youruserslug, setYouruserslug] = createSignal('');
  const [failedalerts, setFailedalerts] = createSignal<AlertMessage[]>([]);
  const [playercount, setPlayercount] = createSignal(0);
  const [loader, setLoader] = createSignal(false);
  const [nameformdone, setNameformdone] = createSignal(false);
  const [firstName, setFirstName] = createSignal('');
  const [formValid, setFormValid] = createSignal(false);

  const [chatScrollRef, setChatScrollRef] = createSignal<HTMLElement | null>(null);
  let formRef: HTMLFormElement | undefined;

  const validateName = (name: string): boolean => {
    const trimmedName = name.trim();
    return trimmedName.length > 3 && trimmedName.length < 10;
  };

  const handleNewMessage = (event: MessageEvent) => {
    const data = event.data.split(/\r?\n/);

    for (let i = 0; i < data.length; i++) {
      const msg: WebSocketMessage = JSON.parse(data[i]);

      // Handle different message types
      if (msg.action === "join-room-notify" && msg.target === roomname()) {
        // Handle join room notification
      }
      
      if (msg.action === "know-yourself" && msg.target === roomname()) {
        setYouruserslug(msg.sender?.slug || '');
      }
      
      if (msg.action === "fail-join-room-notify" && msg.target === roomname()) {
        setGotthrownout(true);
        setFailedalerts(prev => [...prev, { message: msg.message as string }]);
      }
      
      if (msg.action === "found-random-room-notify") {
        setRoomname(msg.message as string);
      }
      
      if (msg.action === "is-room-maker" && msg.target === roomname()) {
        setIsroomaker(true);
      }
      
      if (msg.action === "client-list-notify" && msg.target === roomname()) {
        if (msg.clientlist) {
          setPlayercount(msg.clientlist.length);
          
          const updatedUsers = msg.clientlist.map(element => ({
            username: element.name,
            userslug: element.slug,
            color: element.clientgamemetadata.color,
            score: element.clientgamemetadata.score
          }));
          
          setUsers(updatedUsers);
        }
      }
      
      if (msg.action === "room-bot-greetings" && msg.target === roomname()) {
        setIsthegamestarted(true);
        setGamestarttimer(180);
        gameCountDownTimer();
      }
      
      if (msg.action === "room-bot-end-game" && msg.target === roomname()) {
        if (typeof msg.message === 'string') {
          setWordslist(msg.message.split(','));
        }
        
        if (msg.clientstats) {
          const updatedStats = msg.clientstats.map(element => ({
            username: element.name,
            userslug: element.slug,
            color: element.clientgamemetadata.color,
            score: element.clientgamemetadata.score
          }));
          
          setUserstats(updatedStats);
        }
        
        setIsthegamestoped(true);
      }
      
      if (msg.action === "send-message" && msg.target === roomname() && msg.sender) {
        let newMessage: GameMessage;
        
        if (msg.sender.name === "bot-of-the-room") {
          newMessage = {
            user: "bot",
            useravatar: "🤖",
            message: msg.message as string,
            color: "purple darken-4"
          };
        } else if (msg.sender.slug === youruserslug()) {
          newMessage = {
            user: "you",
            useravatar: msg.sender.name.substring(0, 2),
            message: msg.message as string,
            color: msg.sender.clientgamemetadata?.color || ''
          };
        } else {
          newMessage = {
            user: "other",
            useravatar: msg.sender.name.substring(0, 2),
            message: msg.message as string,
            color: msg.sender.clientgamemetadata?.color || ''
          };
        }
        
        setGamemessages(prev => [...prev, newMessage]);
      }
      
      if (msg.action === "message-by-bot" && msg.target === roomname()) {
        const newMessage: GameMessage = {
          user: "bot",
          useravatar: "🤖",
          message: msg.message as string,
          color: "purple darken-4"
        };
        
        setGamemessages(prev => [...prev, newMessage]);
        
        if (msg.clientstats) {
          setUsers(prev => {
            const newUsers = [...prev];
            for (let i = 0; i < msg.clientstats!.length; i++) {
              if (newUsers[i] && newUsers[i].userslug === msg.clientstats![i].slug) {
                newUsers[i].score = msg.clientstats![i].clientgamemetadata.score;
              }
            }
            return newUsers;
          });
        }
        
        if (msg.letter !== undefined && msg.letter !== "") {
          setNewletter(msg.letter);
        }
        
        if (msg.whichclientturn) {
          if (msg.whichclientturn.slug === youruserslug()) {
            if (gameticker()) {
              clearTimeout(gameticker()!);
            }
            
            if (msg.timer !== undefined) {
              setUsercanentermessagetimer(msg.timer);
              inGamecountDownTimer();
            }
          } else {
            setUsercanentermessagetimer(null);
            setUsercanentermessage(false);
          }
        } else if (msg.timer !== undefined) {
          setUsercanentermessagetimer(msg.timer);
          inGamecountDownTimer();
        }
      }

      // Scroll chat to bottom after message processing
      onMount(() => {
        setTimeout(() => {
          const el = chatScrollRef();
          if (el) {
            el.scrollTop = el.scrollHeight; // ✅ Safe to scroll
          }
        }, 0);
      });
    }
  };

  const connectToWebsocket = () => {
    const websocket = new WebSocket("ws://localhost:8080/wsmesh");
    websocket.onopen = () => console.log("Connected!");
    websocket.onerror = (e) => console.log("WebSocket Error", e);
    websocket.onmessage = (msg) => console.log("Message Received", msg.data);
    websocket.addEventListener('message', handleNewMessage);
    setWs(websocket);
  };

  const waitForSocketConnection = (callback: () => void) => {
    setLoader(true);
    const socket = ws();
    
    if (!socket) return;
    
    if (socket.readyState === 1) {
      setLoader(false);
      callback();
      return;
    }
    
    setTimeout(() => waitForSocketConnection(callback), 5);
  };

  const connectToGameRoom = () => {
    const currentRoomname = roomname();

    if (currentRoomname !== null) {
      if (currentRoomname.length > 10) {
        alert("Roomname length should be equal to 10");
      } else {
        const format = /[ `!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]/;
        if (!format.test(currentRoomname)) {
          waitForSocketConnection(() => {
            ws()?.send(JSON.stringify({ action: 'join-room', message: currentRoomname }));
          });
        } else {
          alert("Roomname not valid");
        }
      }
    } else {
      setIsrandomgame(true);
      waitForSocketConnection(() => {
        ws()?.send(JSON.stringify({ action: 'join-random-room', message: makeroom(10) }));
      });
    }
  };

  const makeroom = (length: number): string => {
    console.log("MAKE A ROOM")
    let result:string = '';
    const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    const charactersLength = characters.length;
    let counter = 0;
    
    while (counter < length) {
      result += characters.charAt(Math.floor(Math.random() * charactersLength));
      counter += 1;
    }
    
    return result;
  };

  const startTheGame = () => {
    waitForSocketConnection(() => {
      ws()?.send(JSON.stringify({ action: 'start-the-game', message: roomname() }));
    });
    
    setStartthegamebuttonloader(true);
  };

  const leaveTheRoom = () => {
    waitForSocketConnection(() => {
      ws()?.send(JSON.stringify({ action: 'leave-room', message: roomname() }));
    });
    
    navigate(-1);
  };

  const sendNewName = (e: Event) => {
    e.preventDefault();
    
    if (validateName(firstName())) {
      waitForSocketConnection(() => {
        ws()?.send(JSON.stringify({ action: 'client-name', message: firstName() }));
      });
      
      setNameformdone(true);
      connectToGameRoom();
    }
  };

  const notSendNewName = () => {
    setNameformdone(true);
    connectToGameRoom();
  };

  const sendinGameMessage = () => {
    const message = ingamemessage();
    
    if (message !== "") {
      if (newletter() !== "") {
        setUsercanentermessage(false);
        setUsercanentermessagetimer(null);
        setUserguessedbeforeticker(true);
        
        waitForSocketConnection(() => {
          ws()?.send(JSON.stringify({
            action: "send-message-by-bot",
            target: roomname(),
            message: message.trim()
          }));
          
          setIngamemessage("");
        });
      } else {
        waitForSocketConnection(() => {
          ws()?.send(JSON.stringify({
            action: "send-message",
            target: roomname(),
            message: message
          }));
          
          setIngamemessage("");
        });
      }
    }
  };

  const gameCountDownTimer = () => {
    const timer = gamestarttimer();
    
    if (timer !== null && timer > 0) {
      setTimeout(() => {
        setGamestarttimer(timer - 1);
        setGameloaderticker(prev => prev + 0.555);
        
        if (timer <= 90 && timer > 30) {
          setGameloadercolor("yellow");
        }
        
        if (timer <= 30) {
          setGameloadercolor("red");
        }
        
        gameCountDownTimer();
      }, 1000);
    }
  };

  const inGamecountDownTimer = () => {
    setUsercanentermessage(true);
    const timer = usercanentermessagetimer();
    
    if (timer !== null && timer > 0) {
      const timeout = setTimeout(() => {
        setUsercanentermessagetimer(timer - 1);
        inGamecountDownTimer();
      }, 1000);
      
      setGameticker(timeout);
    } else {
      setUsercanentermessage(false);
      setUsercanentermessagetimer(null);
      
      if (newletter() !== "") {
        setUserguessedbeforeticker(false);
      }
    }
  };

  const copyText = () => {
    setCopytoclipboard(true);
    const textToCopy = `${url()}${location.pathname}${location.search}`;
    navigator.clipboard.writeText(textToCopy);
  };

  // Lifecycle hooks
  onMount(() => {
    connectToWebsocket();
    
    // Add GSAP animations
    gsap.from('.fade-in', { 
      opacity: 0, 
      y: 20, 
      duration: 0.8, 
      stagger: 0.2,
      ease: 'power2.out' 
    });
  });

  onCleanup(() => {
    const currentRoomname = roomname();
    
    if (currentRoomname !== null && nameformdone()) {
      waitForSocketConnection(() => {
        ws()?.send(JSON.stringify({ action: 'leave-room', message: currentRoomname }));
      });
    }
    
    ws()?.close();
    
    if (gameticker()) {
      clearTimeout(gameticker()!);
    }
  });

  // Effects
  createEffect(() => {
    setFormValid(validateName(firstName()));
  });

  return (
    <Container class="pb-0 fade-in">
      <Show when={loader()}>
        <div class="text-center flex justify-center items-center">
          <CircularProgress color="secondary" />
        </div>
        <div class="text-center flex justify-center items-center">
          <p>Initializing a room</p>
        </div>
      </Show>

      <Show when={!loader() && !nameformdone()}>
        <PlayerDetails
          firstName={firstName} 
          setFirstName={setFirstName} 
          sendNewName={sendNewName} 
          notSendNewName={notSendNewName}
          formValid={formValid}
        />
      </Show>

      <Show when={!loader() && nameformdone() && !isthegamestarted() && roomname() !== null}>
        <PlayerLobby
          users={users}
          failedalerts={failedalerts}
          setFailedalerts={setFailedalerts}
          playercount={playercount}
          gotthrownout={gotthrownout}
          israndomgame={israndomgame}
          copyText={copyText}
          url={url}
          isroomaker={isroomaker}
          startthegamebuttonloader={startthegamebuttonloader}
          startTheGame={startTheGame}
        />
      </Show>

      <Show when={!loader() && nameformdone() && isthegamestarted() && !isthegamestoped() && roomname() !== null}>
        <div class="text-center flex flex-col pb-0 fade-in" style={{ "height": "auto" }}>
          <Container style={{ "max-width": "500px", "height": "auto" }} class="mt-0 pt-0 pb-0 mb-0">
            <div class="flex items-start" style={{ "height": "auto" }}>
              <div class="flex items-center w-full">
                <Card class="mx-auto pb-0 mb-0 w-full">
                  <div class="flex flex-row justify-center">
                    <For each={users()}>
                      {(user) => (
                        <Chip  icon={<Avatar style={{ "background-color": user.color }}>
                        {user.score}
                      </Avatar>} label={user.username.substring(0, 2)} variant="filled" class="mr-2"/>
                      )}
                    </For>
                  </div>
                </Card>
              </div>
            </div>

            <div class="flex items-center mt-0 pt-0" style={{ "height": "auto" }}>
              <div class="flex items-center w-full">
                <LinearProgress
                  variant="determinate"
                  value={gameloaderticker()}
                //   color={gameloadercolor()}
                />
                <Card class="mx-auto w-full" variant="outlined">
                  <div class="flex items-end" style={{ "height": "65vh", "overflow": "auto" }} ref={chatScrollRef}>
                    <div class="w-full">
                      <For each={gamemessages()}>
                        {(message, index) => (
                          <div class="flex flex-row justify-end p-2">
                            <Show when={message.user === 'you'}>
                              <Card
                                class="ml-auto rounded-bs-xl mr-3"
                                style={{
                                  "max-width": "344px",
                                  "background-color": message.color
                                }}
                              >
                                <CardContent>
                                  <span class="text-blue" innerHTML={message.message} />
                                </CardContent>
                              </Card>
                            </Show>

                            <Avatar style={{ "background-color": message.color }}>
                              {message.useravatar}
                            </Avatar>

                            <Show when={message.user !== 'you'}>
                              <Card
                                class="mr-auto rounded-be-xl ml-3"
                                style={{
                                  "max-width": "344px",
                                  "background-color": message.color
                                }}
                              >
                                <CardContent>
                                  <span class="text-blue" innerHTML={message.message} />
                                </CardContent>
                              </Card>
                            </Show>
                          </div>
                        )}
                      </For>
                    </div>
                  </div>
                </Card>
              </div>
            </div>

            <div class="flex items-end my-0 py-0" style={{ "height": "auto" }}>
              <div class="flex items-center mr-0 pr-0 w-full">
                <Card class="mx-auto pb-0 mb-0 w-full">
                  <div class="flex flex-row items-center">
                    <TextField
                      label="Enter your word"
                      style={{ "min-height": "auto" }}
                      class="pt-1 pb-0 mb-0 pr-2 flex-grow"
                      variant="outlined"
                      size="small"
                      value={ingamemessage()}
                      onChange={(e) => setIngamemessage(e.target.value)}
                      disabled={!usercanentermessage()}
                      onKeyDown={(e) => e.key === 'Enter' && sendinGameMessage()}
                    />
                    <Button
                      disabled={!usercanentermessage()}
                      class="px-2 mr-2"
                      variant="contained"
                      color="primary"
                      onClick={sendinGameMessage}
                    >
                      <span class="mdi mdi-send"></span>
                    </Button>
                    <Show when={usercanentermessage()}>
                      <Avatar sx={{ bgcolor: deepOrange[500] }}>
                        {usercanentermessagetimer()}
                      </Avatar>
                    </Show>
                  </div>
                </Card>
              </div>
            </div>
          </Container>
        </div>
      </Show>

      <Show when={!loader() && nameformdone() && isthegamestarted() && isthegamestoped()}>
        <PlayerLeaderboard
          userstats={userstats}
          wordslist={wordslist}
          leaveTheRoom={leaveTheRoom}
        />
      </Show>
    </Container>
  );
};

export default PlayClashofWords;