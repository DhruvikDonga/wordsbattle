<template>
    <div>
       <video ref="videoJsPlayer" class="video-js vjs-defaultskin"></video>
    </div>
 </template>
 <script>
    // Importing video-js
    import videojs from 'video.js';
    export default {
       name: 'VideoJsPlayer',
       props: {
          options: {
             type: Object,
             default() {
                return {};
             }
          }
       },
       data() {
          return {
          player: null
       }
    },
    // initializing the video player
    // when the component is being mounted
    mounted() {
       this.player = videojs(this.$refs.videoPlayer, this.options,() => {
          this.player.log('video player ready', this);
       });
    },
    // destroying the video player
    // when the component is being destroyed
    beforeUnmount() {
       if (this.player) {
          this.player.dispose();
       }
    }
 }
 </script>