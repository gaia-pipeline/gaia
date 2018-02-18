<template>
  <div class="columns is-multiline">
    <template v-for="(pipeline, index) in pipelines">
      <div class="column is-one-third" :key="index">
        <div class="notification content-article">
          <div class="status-display-success"></div>
          <div class="outer-box">
            <div class="outer-box-icon-image">
              <img :src="getImagePath(pipeline.type)" class="outer-box-image">
            </div>
            <div>
              <router-link :to="{ path: '/pipelines/detail', query: { pipelineid: pipeline.id }}" class="subtitle">{{ pipeline.name }}</router-link>
            </div>
            <div>
              <hr style="color: lightgrey;">
              <a class="button is-primary" @click="startPipeline(pipeline.id)">
                <span class="icon">
                  <i class="fa fa-play-circle"></i>
                </span>
                <span>Start Pipeline</span>
              </a>
            </div>
          </div>
        </div>
      </div>
    </template>          
  </div>
</template>

<script>
export default {
  data () {
    return {
      pipelines: []
    }
  },

  mounted () {
    // Fetch data from backend
    this.fetchData()

    // periodically update dashboard
    setInterval(function () {
      this.fetchData()
    }.bind(this), 3000)
  },

  watch: {
    '$route': 'fetchData'
  },

  methods: {
    fetchData () {
      this.$http
        .get('/api/v1/pipelines', { showProgressBar: false })
        .then(response => {
          if (response.data) {
            this.pipelines = response.data
          }
        })
        .catch(error => {
          console.log(error.response.data)
        })
    },

    startPipeline (pipelineid) {
      // Send start request
      this.$http
        .get('/api/v1/pipelines/start/' + pipelineid)
        .then(response => {
          this.$router.push({path: '/pipelines/detail', query: { pipelineid: pipelineid }})
        })
        .catch(error => {
          console.log(error.response.data)
        })
    },

    getImagePath (type) {
      return require('assets/' + type + '.png')
    }
  }
}
</script>

<style lang="scss">

@mixin status-display {
  position: fixed;
  min-width: 50px;
  height: 100%;
  margin-left: -23px;
  margin-top: -20px;
  margin-bottom: -20px;
  border-top-left-radius: 3px;
  border-bottom-left-radius: 3px;
  margin-right: 10px;
}

.status-display-success {
  @include status-display();
  background-color: rgb(49, 196, 49);  
}

.status-display-folder {
  @include status-display();
  background-color: #4da2fc;
}

.status-display-fail {
  @include status-display();
  background-color: #ca280b;
}

.outer-box {
  padding-left: 40px; 
  min-height: 170px;
  width: 100%;
}

.outer-box-icon {
  width: 50px;
  float: left;
}

.outer-box-icon-image {
  float: left;
  width: 40px;
  height: 40px;
  overflow: hidden;
  border-radius: 50%;
  position: relative;
  border-color: whitesmoke;
  border-style: solid;
  margin-right: 10px;
  margin-top: -5px;
}

.outer-box-image {
  position: absolute;
  width: 50px;
  height: 50px;
  top: 70%;
  left: 50%;
  transform: translate(-50%, -50%);
}

</style>
