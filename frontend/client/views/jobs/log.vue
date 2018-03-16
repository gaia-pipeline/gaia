<template>
  <div class="job-log-view">
    <message :direction="'down'" :message="'This is a cool test and test and test and test and test and test'" :duration="0"></message>
    <div class="job-loading" v-if="jobRunning"></div>
  </div>
</template>

<script>
import Message from 'vue-bulma-message'

export default {
  data () {
    return {
      job: null,
      jobRunning: true,
      runID: null,
      pipelineID: null
    }
  },

  mounted () {
    // Fetch data from backend
    this.fetchData()

    // periodically update dashboard
    this.intervalID = setInterval(function () {
      this.fetchData()
    }.bind(this), 3000)
  },

  watch: {
    '$route': 'fetchData'
  },

  components: {
    Message
  },

  methods: {
    fetchData () {
      // look up url parameters
      this.pipelineID = this.$route.query.pipelineid
      this.runID = this.$route.query.runid
      if (!this.runID || !this.pipelineID) {
        return
      }

      this.$http
        .get('/api/v1/pipelines', { showProgressBar: false })
        .then(response => {
          if (response.data) {
            this.pipelines = response.data
          }
        })
        .catch((error) => {
          clearInterval(this.intervalID)
          this.$onError(error)
        })
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
