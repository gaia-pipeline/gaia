<template>
  <div class="columns is-multiline">
    <template v-for="(pipeline, index) in pipelines">
      <div class="column is-one-third" :key="index">
        <div class="notification content-article">
          <div class="status-display-success"></div>
          <div class="outer-box">
            <router-link :to="{ path: '/pipeline/detail', query: { pipelineid: pipeline.p.id }}" class="hoveraction">
              <div class="outer-box-icon-image">
                <img :src="getImagePath(pipeline.p.type)" class="outer-box-image">
              </div>
              <div>
                <span class="subtitle">{{ pipeline.p.name }}</span>
              </div>
            </router-link>            
            <hr class="pipeline-hr">
            <div class="pipeline-info">
              <span>Duration: {{ pipeline.r.startdate }}</span><br />
              <span>Started: {{ pipeline.r.finishdate }}</span><br />
              <div class="pipelinegrid-footer">
                <a class="button is-primary" @click="startPipeline(pipeline.p.id)" style="width: 250px;">
                  <span class="icon">
                    <i class="fa fa-play-circle"></i>
                  </span>
                  <span>Start Pipeline</span>
                </a>
              </div>
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
    var intervalID = setInterval(function () {
      this.fetchData()
    }.bind(this), 3000)

    // Append interval id to store
    this.$store.commit('appendInterval', intervalID)
  },

  watch: {
    '$route': 'fetchData'
  },

  methods: {
    fetchData () {
      this.$http
        .get('/api/v1/pipeline/latest', { showProgressBar: false })
        .then(response => {
          if (response.data) {
            this.pipelines = response.data
          }
        })
        .catch((error) => {
          this.$store.commit('clearIntervals')
          this.$onError(error)
        })
    },

    startPipeline (pipelineid) {
      // Send start request
      this.$http
        .post('/api/v1/pipeline/' + pipelineid + '/start')
        .then(response => {
          if (response.data) {
            this.$router.push({path: '/pipeline/detail', query: { pipelineid: pipelineid, runid: response.data.id }})
          }
        })
        .catch((error) => {
          this.$store.commit('clearIntervals')
          this.$onError(error)
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

.hoveraction:hover .outer-box-icon-image {
  border-color: #4da2fc !important;
}

.hoveraction:hover .subtitle {
  color: #4da2fc !important;
  text-decoration: underline;
  text-decoration-color: #4da2fc !important;
}

.pipeline-hr {
  position: absolute;
  width: 325px;
  margin-left: -13px;
  margin-top: 18px;
  background-image: linear-gradient(
    to right,
    black 33%,
    rgba(255, 255, 255, 0) 0%
  );
  background-position: bottom;
  background-size: 3px 1px;
  background-repeat: repeat-x;
}

.pipeline-info {
  padding-top: 33px;
}

.pipelinegrid-footer {
  margin: auto;
  width: 82%;
  padding-top: 20px;
}

</style>
