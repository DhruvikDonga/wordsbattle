
import { Transition } from 'vue';

<template>
  <v-app>
   <v-app-bar color="purple darken-4" class="text-center"  app>
      <v-toolbar-title>
        Clash of Words <span style="font-size:x-small;font-stretch:extra-expanded;">v1.0.0</span>
      </v-toolbar-title>
      <div class="pr-1">
        <v-btn density="compact" icon="mdi-github"></v-btn>
      
      </div>
   </v-app-bar>
    <v-main>
      <v-container fluid >
        <!-- <nav>
          <router-link to="/">Home</router-link> |
          <router-link to="/about">About</router-link>
        </nav> -->
        <router-view v-slot="{ Component }">
          <Transition mode="out-in">
            <keep-alive exclude="PlaywithFriend">
              <component :is="Component" />
            </keep-alive>

          </Transition>

        </router-view>
        <!-- <router-view v-slot="{ Component }">
          <Transition mode="out-in">
            <keep-alive exclude="PlaywithFriendGM">
              <component :is="Component" />
            </keep-alive>

          </Transition>

        </router-view> -->
        </v-container>
      </v-main>
    <v-footer max-height="200"  app class="d-flex justify-start ">
      <!-- <small>wordsbattle@</small>
    <v-icon
      size="small"
      color="dark"
      icon="mdi-github"
    ></v-icon> -->
    </v-footer>

  </v-app>
</template>


<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
}

nav {
  padding: 30px;
}

nav a {
  font-weight: bold;
  color: #2c3e50;
}

nav a.router-link-exact-active {
  color: #42b983;
}

.v-enter-active,
.v-leave-active {
  transition: opacity 0.5s ease;
}

.v-enter-from,
.v-leave-to {
  opacity: 0;
}
</style>

