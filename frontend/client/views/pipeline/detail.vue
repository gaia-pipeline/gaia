<template>
  <div class="tile is-ancestor">
    <div class="tile is-vertical">
      <div class="tile is-parent">
        <a class="button is-primary" @click="startPipeline(pipelineID)" style="margin-right: 10px;">
          <span class="icon">
            <i class="fa fa-play-circle"></i>
          </span>
          <span>Start Pipeline</span>
        </a>
        <a class="button is-green-button" @click="jobLog" v-if="runID">
          <span class="icon">
            <i class="fa fa-terminal"></i>
          </span>
          <span>Show Logs</span>
        </a>
      </div>

      <div class="tile">
        <div class="tile is-vertical is-parent is-12">
          <article class="tile is-child notification content-article">
            <div id="pipeline-detail"></div>
          </article>
        </div>
      </div>

      <div class="tile is-parent">
        <article class="tile is-child notification content-article box">
            <vue-good-table
              title="Previous runs"
              :columns="runsColumns"
              :rows="runsRows"
              :paginate="true"
              :global-search="true"
              :defaultSortBy="{field: 'id', type: 'desc'}"
              globalSearchPlaceholder="Search ..."
              styleClass="table table-own-bordered">
              <template slot="table-row" slot-scope="props">
                <td>
                  <router-link :to="{ path: '/pipeline/detail', query: { pipelineid: pipelineID, runid: props.row.id }}" class="is-blue">
                    {{ props.row.id }}
                  </router-link>
                </td>
                <td>
                  <span v-if="props.row.status === 'success'" style="color: green;">{{ props.row.status }}</span>
                  <span v-else-if="props.row.status === 'failed'" style="color: red;">{{ props.row.status }}</span>
                  <span v-else>{{ props.row.status }}</span>
                </td>
                <td>{{ calculateDuration(props.row.startdate, props.row.finishdate) }}</td>
              </template>
              <div slot="emptystate" class="empty-table-text">
                No pipeline runs found in database.
              </div>
            </vue-good-table>
        </article>
      </div>
    </div>
  </div>
</template>

<script>
import Vue from 'vue'
import Vis from 'vis'
import VueGoodTable from 'vue-good-table'
import moment from 'moment'

Vue.use(VueGoodTable)

export default {

  data () {
    return {
      pipelineID: null,
      runID: null,
      nodes: null,
      edges: null,
      lastRedraw: false,
      runsColumns: [
        {
          label: 'ID',
          field: 'id',
          type: 'number'
        },
        {
          label: 'Status',
          field: 'status'
        },
        {
          label: 'Duration'
        }
      ],
      runsRows: [],
      pipelineViewOptions: {
        physics: { stabilization: true },
        layout: {
          hierarchical: {
            enabled: true,
            levelSeparation: 200,
            direction: 'LR',
            sortMethod: 'directed'
          }
        },
        nodes: {
          borderWidth: 4,
          size: 40,
          color: {
            border: '#222222'
          },
          font: { color: '#eeeeee' }
        },
        edges: {
          smooth: {
            type: 'cubicBezier',
            forceDirection: 'vertical',
            roundness: 0.4
          },
          color: {
            color: 'whitesmoke',
            highlight: '#4da2fc'
          },
          arrows: {to: true}
        }
      }
    }
  },

  mounted () {
    // View should be re-rendered
    this.lastRedraw = false

    // periodically update view
    this.fetchData()
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
    '$route': 'locationReload'
  },

  methods: {
    locationReload () {
      // View should be re-rendered
      this.lastRedraw = false

      // Fetch data
      this.fetchData()
    },

    fetchData () {
      // look up url parameters
      var pipelineID = this.$route.query.pipelineid
      if (!pipelineID) {
        return
      }
      this.pipelineID = pipelineID

      // runID is optional
      var runID = this.$route.query.runid

      // If runid was set, look up this run
      if (runID) {
        // set run id
        this.runID = runID

        // Run ID specified. Do concurrent request
        this.$http.all([this.getPipeline(pipelineID), this.getPipelineRun(pipelineID, runID), this.getPipelineRuns(pipelineID)])
          .then(this.$http.spread(function (pipeline, pipelineRun, pipelineRuns) {
            // We only redraw the pipeline if pipeline is running
            if (pipelineRun.data.status !== 'running' && !this.lastRedraw) {
              this.drawPipelineDetail(pipeline.data, pipelineRun.data)
              this.lastRedraw = true
            } else if (pipelineRun.data.status === 'running') {
              this.lastRedraw = false
              this.drawPipelineDetail(pipeline.data, pipelineRun.data)
            }
            this.runsRows = pipelineRuns.data
          }.bind(this)))
          .catch((error) => {
            this.$store.commit('clearIntervals')
            this.$onError(error)
          })
      } else {
        // Do concurrent request
        this.$http.all([this.getPipeline(pipelineID), this.getPipelineRuns(pipelineID)])
          .then(this.$http.spread(function (pipeline, pipelineRuns) {
            if (!this.lastRedraw) {
              this.drawPipelineDetail(pipeline.data, null)
              this.lastRedraw = true
            }

            // Are runs available?
            if (pipelineRuns.data) {
              this.runsRows = pipelineRuns.data
            }
          }.bind(this)))
          .catch((error) => {
            this.$store.commit('clearIntervals')
            this.$onError(error)
          })
      }
    },

    getPipeline (pipelineID) {
      return this.$http.get('/api/v1/pipeline/' + pipelineID, { showProgressBar: false })
    },

    getPipelineRun (pipelineID, runID) {
      return this.$http.get('/api/v1/pipelinerun/' + pipelineID + '/' + runID, { showProgressBar: false })
    },

    getPipelineRuns (pipelineID) {
      return this.$http.get('/api/v1/pipelinerun/' + pipelineID, { showProgressBar: false })
    },

    drawPipelineDetail (pipeline, pipelineRun) {
      // Check if pipelineRun was set
      var jobs = null
      if (pipelineRun) {
        jobs = pipelineRun.jobs
      } else {
        jobs = pipeline.jobs
      }

      // Check if something has changed
      if (this.nodes) {
        var redraw = false
        for (let i = 0, l = this.nodes.length; i < l; i++) {
          for (let x = 0, y = jobs.length; x < y; x++) {
            if (this.nodes._data[i].internalID === jobs[x].id && this.nodes._data[i].internalStatus !== jobs[x].status) {
              redraw = true
              break
            }
          }
        }

        // Check if we have to redraw
        if (!redraw) {
          return
        }
      }

      // Initiate data structure
      var nodesArray = []
      var edgesArray = []

      // Iterate all jobs of the pipeline
      for (let i = 0, l = jobs.length; i < l; i++) {
        // Choose the image for this node
        var nodeImage = require('assets/questionmark.png')
        if (jobs[i].status) {
          switch (jobs[i].status) {
            case 'success':
              nodeImage = require('assets/success.png')
              break
            case 'failed':
              nodeImage = require('assets/fail.png')
              break
          }
        }

        // Create nodes object
        let node = {
          id: i,
          internalID: jobs[i].id,
          internalStatus: jobs[i].status,
          shape: 'circularImage',
          image: nodeImage,
          label: jobs[i].title,
          font: {
            color: '#eeeeee'
          }
        }

        // Add node to nodes list
        nodesArray.push(node)

        // Iterate all jobs again to find the next highest job priority
        var highestPrio = null
        for (let x = 0, y = jobs.length; x < y; x++) {
          if (jobs[x].priority > jobs[i].priority && (jobs[x].priority < highestPrio || !highestPrio)) {
            highestPrio = jobs[x].priority
          }
        }

        // Iterate again all jobs to set all edges
        if (highestPrio) {
          for (let x = 0, y = jobs.length; x < y; x++) {
            if (jobs[x].priority === highestPrio) {
              // create edge
              let edge = {
                from: i,
                to: x
              }

              // add edge to edges list
              edgesArray.push(edge)
            }
          }
        }
      }

      // If pipelineView already exist, just update it
      if (window.pipelineView && this.nodes && this.edges) {
        // Redraw
        this.nodes.clear()
        this.edges.clear()
        this.nodes.add(nodesArray)
        this.edges.add(edgesArray)
      } else {
        // translate to vis data structure
        this.nodes = new Vis.DataSet(nodesArray)
        this.edges = new Vis.DataSet(edgesArray)

        // prepare data object for vis
        var data = {
          nodes: this.nodes,
          edges: this.edges
        }

        // Find container
        var container = document.getElementById('pipeline-detail')

        // Create vis network
        // We have to move out the instance out of vue because of https://github.com/almende/vis/issues/2567
        window.pipelineView = new Vis.Network(container, data, this.pipelineViewOptions)
      }
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

    jobLog () {
      // Route
      this.$router.push({path: '/pipeline/log', query: { pipelineid: this.pipelineID, runid: this.runID }})
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
    }
  }

}
</script>

<style lang="scss">

#pipeline-detail {
  width: 100%;
  height: 400px;
}

</style>
