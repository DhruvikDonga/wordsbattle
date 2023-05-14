<template>
  <v-container>
    <v-row class="text-center" align="start">
      <v-col cols="12">
        <v-img
          :src="require('../assets/clashofwordslogo.png')"
          class="my-3"
          contain
          height="200"
        />
      </v-col>

      <v-col class="mb-4">
        <h1 class="display-2 mb-3" style="font-family:'Trebuchet MS', 'Lucida Sans Unicode', 'Lucida Grande', 'Lucida Sans', Arial, sans-serif;">
          Welcome to the <span style="color:blueviolet;">Clash</span> <span style="color: yellow;">of</span> <span style="color: red;">Words</span>
        </h1>
        <p class="subheading font-weight-regular">
          Let the battle begin ⚔
        </p>
        <div v-if="active!= null">Current online :- {{ active }}</div>
        <v-container>
          <v-row align="center" justify="center"> 
            <v-col cols="auto">
              <v-btn block rounded primary @click="playwithfriend()" color="blue-darken-4" append-icon="mdi-account-group"  style="font-family: Cambria; text-transform: unset">Play with friends</v-btn>
            </v-col>    
            <v-col cols="auto">
              <v-btn block  rounded @click="playwithrandomfriend()" color="yellow" append-icon="mdi-account-multiple" style="font-family: Cambria; text-transform: unset">Random Game</v-btn>
            </v-col>
            <v-dialog
              v-model="dialog"
              width="auto"
            >
              <template v-slot:activator="{ props }">
                <v-col cols="auto">
                  <v-btn block rounded primary v-bind="props" color="green" append-icon="mdi-help"  style="font-family: Cambria; text-transform: unset">About the game</v-btn>
                </v-col>
              </template>
              <v-card>
                <v-card-text>
                  <h1 style="font-family: Cambria; text-transform: unset ; text-align: center;">About the Game</h1><br>
                  Clash of words is a text based game based on chat server theme where 2 or more players have to simply give a correct word to a letter asked by in game bot.
                  You can play a Random Game or with Friends.
                  <br>
                  <br>
                  <ul>
                    <li>
                      In <b>play with random game</b> you are connected to a random player online and it will automatically start the game.
                      <v-row>
                        <v-col align-self="center">
                          <v-img
                          :width="700"
                          aspect-ratio="16/10"
                          cover
                          src="../assets/random-game.gif"
                          justify-center
                          class="mx-auto"
                        ></v-img>
                        </v-col>
                      </v-row>
                    </li>
                    <br>
                    <li>
                      In <b>play with friends game</b> you can add upto 10 friends by the room link and you can play the game with it.
                      <v-row>
                        <v-col align-self="center">
                          <v-img
                          :width="700"
                          aspect-ratio="16/10"
                          cover
                          src="../assets/with-friend.gif"
                          justify-center
                          class="mx-auto"
                        ></v-img>
                        </v-col>
                      </v-row>
                    </li>
                    <br>
                    <li>
                      This is a demo game played between 2 players and has all the rules covered
                      <v-row>
                        <v-col align-self="center">
                          <!-- <v-card
                          class="mx-auto"
                          max-width="100%"
                          max-height="300"
                          width="100%"
                          > -->
                          <video-player src="require('../assets/game-video.mp4')" style="height: 30vh;" class="vjs-fluid vjs-layout-medium" dataType="video/mp4" controls :options="videoOptions"/>
                            <!-- <video width="400" height="300" :src="require('../assets/game-video.mp4')" class="video-js vjs-defaultskin vjs-16-9" controls>
                            </video> -->
                          <!-- </v-card> -->
                        </v-col>
                      </v-row>
                    </li>
                  </ul>
                </v-card-text>
                <v-card-actions>
                  <v-btn color="primary" block @click="dialog = false">Close</v-btn>
                </v-card-actions>
              </v-card>
            </v-dialog>
          </v-row>
        </v-container>
       
      </v-col>
    </v-row>
    <v-row align="end" class="pb-0">
      <v-container class="pb-0">
          <v-card
            class="mx-auto pb-0 mb-0 rounded-shaped"
            max-width="190"
            variant="plain"

          >
            <v-card-item class="pb-0 px-0" >
                <div>
                    <div class="mb-1">
                      <v-row justify="center">
                        <v-col cols="6" class="mr-0 pr-0">
                          <span><small>by</small></span> <span style="font-family:'Franklin Gothic Medium', 'Arial Narrow', Arial, sans-serif;font-size: large;font-stretch: wider;">dhru</span><span style="color: grey;font-family:'Franklin Gothic Medium', 'Arial Narrow', Arial, sans-serif;font-size: large;font-stretch: expanded;">v!k</span><span> & </span>
                        </v-col>
                        <v-col cols="4" class="ml-0 pl-0">
                          <span><small><v-img :src="require('../assets/mwgames3.png')" contain height="25"/></small></span>
                        </v-col>
                    </v-row>  
                      <v-divider
                          :thickness="0.5"
                          class="border-opacity-25 mx-2"
                        ></v-divider>
                        <span style="font-size: x-small;">
                          <v-btn
                            size="small"
                            density="compact"
                            color="dark"
                            icon="mdi-github"
                            href="https://github.com/DhruvikDonga/" 
                            target="_blank"
                          ></v-btn> |
                          <v-btn
                            size="small"
                            density="compact"
                            color="dark"
                            icon="mdi-linkedin"
                            href="https://www.linkedin.com/in/dhruvik-donga-9363a0172/" 
                            target="_blank"
                          ></v-btn> |
                          <v-btn
                            size="small"
                            density="compact"
                            color="dark"
                            icon="mdi-instagram"
                            href="https://www.instagram.com/dongadhruvik/" 
                            target="_blank"
                          ></v-btn>
                        </span>
                        
                    </div>
                </div>
                </v-card-item>
            </v-card>
        </v-container>
    </v-row>
  </v-container>
</template>


<script>
import router from "../router/index"
import { VideoPlayer } from '@videojs-player/vue'
import 'video.js/dist/video-js.css'
export default {
  data() {
    return {
      loader: false,
      serverUrl: "ws://localhost:8080/ws",
      active:null,
      aboutdialog: false,
      dialog: false,
      videoOptions: {
        autoplay: true,
        controls: true,
        sources: [
          {
            src:require('../assets/game-video.mp4'),
            type: 'video/mp4'
          }
        ]
      }
    }
    },
    components: {
      VideoPlayer
    },
    methods: {
      
      playwithfriend() {
        // `route` is either a string or object
        let roomname = this.makeroom(10)
        router.push({path: '/play', query: {room: roomname}});
      },
      playwithrandomfriend() {
        router.push({path: '/play-random'});
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
      }
    },
    beforeRouteUpdate(to, from, next) {
      next();
    }
}

</script>
