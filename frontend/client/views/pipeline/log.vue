<template>
  <div>
    <a class="button is-green-button" @click="backToDetail" style="margin-bottom: 15px;">
      <span class="icon">
        <i class="fa fa-arrow-circle-left"></i>
      </span>
      <span>Go Back</span>
    </a>
    <div class="job-log-view">
      <message :direction="'down'" :message="logText" :duration="0"></message>
      <div class="job-loading" v-if="jobRunning"></div>
    </div>
  </div>
</template>

<script>
import Message from 'vue-bulma-message-html'

export default {
  data () {
    return {
      logText: '',
      jobRunning: true,
      runID: null,
      pipelineID: null
    }
  },

  mounted () {
    // Reset log text
    this.logText = ''

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

  destroyed () {
    this.$store.commit('clearIntervals')
  },

  components: {
    Message
  },

  methods: {
    fetchData () {
      // look up required url parameters
      this.pipelineID = this.$route.query.pipelineid
      this.runID = this.$route.query.runid
      if (!this.runID || !this.pipelineID) {
        return
      }

      this.$http
        .get('/api/v1/pipelinerun/' + this.pipelineID + '/' + this.runID + '/log', { showProgressBar: false })
        .then(response => {
          if (response.data) {
            // We add the received log
            this.logText = response.data.log

            // LF does not work for HTML. Replace with <br />
            this.logText = this.logText.replace(/\n/g, '<br />')

            // All jobs finished. Stop interval.
            if (response.data.finished) {
              this.jobRunning = false
              clearInterval(this.intervalID)
            }
          }
        })
        .catch((error) => {
          clearInterval(this.intervalID)
          this.$onError(error)
        })
    },

    backToDetail () {
      // Route
      this.$router.push({path: '/pipeline/detail', query: { pipelineid: this.pipelineID, runid: this.runID }})
    }
  }
}
</script>

<style lang="scss">

.job-log-view {
  width: 100%;
}

.message-header {
  background-color: #4da2fc;
}

.message-body {
  background-color: black;
  border: none;
  color: whitesmoke;
}

.job-loading {
  margin-top: 20px;
  border: 5px solid #f3f3f3; /* Light grey */
  border-top: 5px solid #3498db; /* Blue */
  border-radius: 50%;
  width: 70px;
  height: 70px;
  animation: spin 2s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

</style>
