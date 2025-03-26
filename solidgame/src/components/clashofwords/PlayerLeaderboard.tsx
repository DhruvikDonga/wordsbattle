import { For } from "solid-js";
import Container from "@suid/material/Container";
import Card from "@suid/material/Card";
import CardHeader from "@suid/material/CardHeader";
import CardContent from "@suid/material/CardContent";
import Divider from "@suid/material/Divider";
import Table from "@suid/material/Table";
import TableHead from "@suid/material/TableHead";
import TableRow from "@suid/material/TableRow";
import TableCell from "@suid/material/TableCell";
import TableBody from "@suid/material/TableBody";
import Chip from "@suid/material/Chip";
import Button from "@suid/material/Button";

const PlayerLeaderboard = (props: { userstats: any; wordslist: any; leaveTheRoom: () => void }) => {
  return (
    <div class="text-center flex flex-col pb-0 fade-in" style={{ height: "auto" }}>
      <Container>
        <Card class="mx-auto" style={{ "max-width": "400px", "max-height": "600px" }} variant="outlined">
          <CardHeader title="Game board 🏆" />
          <Divider class="border-opacity-25 mx-2" />
          <CardContent>
            <Table stickyHeader>
              <TableHead>
                <TableRow>
                  <TableCell>Player</TableCell>
                  <TableCell>Score</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                <For each={props.userstats()}>
                  {(user, index) => (
                    <TableRow>
                      <TableCell>
                        {user.username} <small>@{user.userslug}</small>
                        {index() === 0 && " 🥇"}
                        {index() === 1 && " 🥈"}
                        {index() === 2 && " 🥉"}
                      </TableCell>
                      <TableCell>{user.score}</TableCell>
                    </TableRow>
                  )}
                </For>
              </TableBody>
            </Table>
          </CardContent>
          <Divider class="border-opacity-25 mx-2" />
          <CardContent>
            <div>
              <h4>Words guessed:</h4>
              <div class="mb-1">
                <For each={props.wordslist()}>
                  {(word) => (
                    <Chip class="ma-2" label={word} />
                  )}
                </For>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card class="mx-auto" style={{ "max-width": "200px" }}>
          <Button fullWidth variant="contained" color="primary" class="mt-3" onClick={props.leaveTheRoom}>
            Leave the room
          </Button>
        </Card>
      </Container>
    </div>
  );
};

export default PlayerLeaderboard;
