import { Alert, Avatar,Button,Card,CardContent,CircularProgress,Container } from "@suid/material";
import { Show, For } from "solid-js";

type User = {
  username: string;
  userslug: string;
  color: string;
};

type AlertType = {
  message: string;
};

type PlayerLobbyProps = {
  users: () => User[];
  failedalerts: () => AlertType[];
  setFailedalerts: (updateFn: (prev: AlertType[]) => AlertType[]) => void;
  playercount: () => number;
  gotthrownout: () => boolean | null;
  israndomgame: () => boolean;
  copyText: () => void;
  url: () => string;
  isroomaker: () => boolean;
  startthegamebuttonloader: () => boolean;
  startTheGame: () => void;
};

const PlayerLobby = ({
  users,
  failedalerts,
  setFailedalerts,
  playercount,
  gotthrownout,
  israndomgame,
  copyText,
  url,
  isroomaker,
  startthegamebuttonloader,
  startTheGame,
}: PlayerLobbyProps) => {
  return (
    <div class="text-center flex flex-col items-start fade-in">
      <Show when={users().length > 0}>
        <Container>
          <div class="flex items-start" style={{ height: "auto" }}>
            <For each={users()}>
              {(user) => (
                <div class="px-0 mx-0 flex items-center">
                  <Avatar style={{ "background-color": user.color }}>
                    {user.username.substring(0, 2)}
                  </Avatar>
                </div>
              )}
            </For>
          </div>
        </Container>
      </Show>

      <For each={failedalerts()}>
        {(alert) => (
          <Alert severity="error" onClose={() => setFailedalerts((prev) => prev.filter((a) => a !== alert))}>
            {alert.message}
          </Alert>
        )}
      </For>

      <div class="mb-4">
        <Show when={playercount() < 10 && !gotthrownout() && !israndomgame()}>
          <Container>
            <h4 class="text-2xl font-bold mb-3">Share this url to your friend till we wait in lobby 😃</h4>
            <h4 class="text-2xl font-bold mb-3">Max 10 players allowed</h4>
            <Card class="mx-auto" style={{ "max-width": "344px" }} variant="outlined">
              <div class="flex justify-end pt-1">
                <Button onClick={copyText} size="small" startIcon="mdi-content-copy">Copy</Button>
              </div>
              <CardContent>
                <div class="text-lg mb-1">
                  {url() + location.pathname + location.search}
                </div>
              </CardContent>
            </Card>
          </Container>
        </Show>

        <Show when={playercount() >= 2 && !israndomgame()}>
          <Container>
            <Card class="mx-auto" style={{ "max-width": "344px" }} variant="outlined">
              <CardContent>
                <div class="text-lg mb-1">
                  <h3>Hold tight we gotta go in lads 🐱‍🏍</h3>
                  <Show when={isroomaker() && !startthegamebuttonloader()}>
                    <small>Your friend is waiting let's start this game 😃</small>
                    <Button fullWidth variant="contained" color="primary" class="mt-3" onClick={startTheGame}>
                      Start the Game 🥁
                    </Button>
                  </Show>
                  <Show when={isroomaker() && startthegamebuttonloader()}>
                    <Button fullWidth variant="outlined" color="primary" class="mt-3" disabled>
                      <CircularProgress color="inherit" /> &nbsp;We are into it
                    </Button>
                  </Show>
                  <Show when={!isroomaker()}>
                    <small>Your friend will start the game so hold tight</small>
                  </Show>
                </div>
              </CardContent>
            </Card>
          </Container>
        </Show>
      </div>
    </div>
  );
};

export default PlayerLobby;
