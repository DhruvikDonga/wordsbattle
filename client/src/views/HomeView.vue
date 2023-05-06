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
              <v-btn block  rounded @click="playwithrandomfriend()" color="yellow" append-icon="mdi-account-multiple" style="font-family: Cambria; text-transform: unset">Random Game</v-btn>
            </v-col>
            <v-col cols="auto">
              <v-btn block rounded primary @click="playwithfriend()" color="blue-darken-4" append-icon="mdi-account-group"  style="font-family: Cambria; text-transform: unset">Play with friends</v-btn>
            </v-col>
          </v-row>
        </v-container>
       
      </v-col>
    </v-row>
    <v-row align="end" class="pb-0">
      <v-container class="pb-0">
          <v-card
            class="mx-auto pb-0 mb-0 rounded-shaped"
            max-width="130"
            variant="plain"

          >
            <v-card-item class="pb-0 px-0" >
                <div>
                    <div class="mb-1">
                     
                      <span><small>games by</small></span> <span style="font-family:'Franklin Gothic Medium', 'Arial Narrow', Arial, sans-serif;font-size: large;font-stretch: wider;">dhru</span><span style="color: grey;font-family:'Franklin Gothic Medium', 'Arial Narrow', Arial, sans-serif;font-size: large;font-stretch: expanded;">v!k</span>
                        <br>
                        <v-divider
                          :thickness="0.5"
                          class="border-opacity-25 mx-2"
                        ></v-divider>
                        <span style="font-size: x-small;">
                          <v-icon
                            size="small"
                            color="dark"
                            icon="mdi-github"
                          ></v-icon> |
                          <v-icon
                            size="small"
                            color="dark"
                            icon="mdi-linkedin"
                          ></v-icon> |
                          <v-icon
                            size="small"
                            color="dark"
                            icon="mdi-instagram"
                          ></v-icon>
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
//import ws from "../websocket"
export default {
  data() {
    return {
      loader: false,
      serverUrl: "ws://localhost:8080/ws",
      active:null,
    }
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
