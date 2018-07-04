<template>
  <div class="columns is-multiline">
    <template v-for="(pipeline, index) in pipelines">
      <div class="column is-one-third" :key="index">
        <div class="pipeline-box notification content-article">
          <div class="status-display-success" v-if="pipeline.r.status === 'success'"></div>
          <div class="status-display-fail" v-else-if="pipeline.r.status === 'failed'"></div>
          <div class="status-display-unknown" v-else></div>
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
              <i class="fa fa-hourglass"></i>
              <span style="color: #b1adad;"> 
                Duration: 
              </span>
              <span v-if="pipeline.r.status === 'success' || pipeline.r.status === 'failed'"> 
                {{ calculateDuration(pipeline.r.startdate, pipeline.r.finishdate) }}
              </span>
              <span v-else>
                unknown
              </span><br />
              <i class="fa fa-calendar"></i>
              <span style="color: #b1adad;"> 
                Started: 
              </span>
              <span v-if="pipeline.r.status === 'success' || pipeline.r.status === 'failed'"> 
                {{ humanizedDate(pipeline.r.finishdate) }}
              </span>
              <span v-else>
                unknown
              </span><br />
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
    <div v-if="pipelines.length == 0" class="no-pipelines-div">
      <span class="no-pipelines-text">No pipelines are available. Please create a pipeline first.</span>
    </div>         
  </div>
</template>

<script>
import moment from 'moment'

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

  destroyed () {
    this.$store.commit('clearIntervals')
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
    },

    calculateDuration (startdate, finishdate) {
      if (!moment(startdate).millisecond()) {
        startdate = moment()
      }
      if (!moment(finishdate).millisecond()) {
        finishdate = moment()
      }

      // Calculate difference
      var diff = moment(finishdate).diff(moment(startdate), 'seconds')
      if (diff < 60) {
        return diff + ' seconds'
      }
      return moment.duration(diff, 'seconds').humanize()
    },

    humanizedDate (date) {
      return moment(date).format('LLL')
    }
  }
}
</script>

<style lang="scss">

.no-pipelines-div {
  width: 100%;
  text-align: center;
  margin-top: 50px;
}

.no-pipelines-text {
  color: whitesmoke;
  font-size: 2rem;
}

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

.status-display-unknown {
  @include status-display();
  background-color: grey;
}

.pipeline-box {
  width: 377px;
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
