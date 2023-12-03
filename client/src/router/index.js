import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import PlaywithFriend from '../views/PlaywithFriend.vue'
import PlaywithFriendGM from '../views/PlaywithFriendGM.vue'

const routes = [
  {
    path: '/',
    name: 'home',
    component: HomeView
  },
  {
    path: '/play',
    name: 'room',
    component: PlaywithFriend,
    meta: { transition: 'slide-left' },

  },
  {
    path: '/play-random',
    name: 'randomroom',
    component: PlaywithFriend,
    meta: { transition: 'slide-left' },

  },
  {
    path: '/play-gm',
    name: 'roomgm',
    component: PlaywithFriendGM,
    meta: { transition: 'slide-left' },

  },
  {
    path: '/play-random-gm',
    name: 'randomroomgm',
    component: PlaywithFriendGM,
    meta: { transition: 'slide-left' },

  },
  {
    path: '/about',
    name: 'about',
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () => import(/* webpackChunkName: "about" */ '../views/AboutView.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
