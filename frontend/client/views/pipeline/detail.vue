<template>
  <div class="tile is-ancestor">
    <div class="tile is-vertical">
      <div class="tile is-parent">
        <a class="button is-primary" @click="checkPipelineArgs" style="margin-right: 10px;">
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
                <td>
                  <a v-on:click="stopPipelineModal(pipelineID, props.row.id)"><i class="fa fa-ban" style="color: whitesmoke;"></i></a>
                </td>
              </template>
              <div slot="emptystate" class="empty-table-text">
                No pipeline runs found in database.
              </div>
            </vue-good-table>
        </article>

        <!-- stop pipeline run modal -->
        <modal :visible="showStopPipelineModal" class="modal-z-index" @close="close">
          <div class="box stop-pipeline-modal">
            <article class="media">
              <div class="media-content">
                <div class="content">
                  <p>
                    <span style="color: white;">Do you really want to cancel this run?</span>
                  </p>
                </div>
                <div class="modal-footer">
                  <div style="float: left;">
                    <button class="button is-primary" v-on:click="stopPipeline" style="width:150px;">Yes</button>
                  </div>
                  <div style="float: right;">
                    <button class="button is-danger" v-on:click="close" style="width:130px;">No</button>
                  </div>
                </div>
              </div>
            </article>
          </div>
        </modal>
      </div>

    </div>
  </div>
</template>

<script>
import Vue from 'vue'
import Vis from 'vis'
import { Modal } from 'vue-bulma-modal'
import VueGoodTable from 'vue-good-table'
import moment from 'moment'

Vue.use(VueGoodTable)

export default {
  components: {
    Modal
  },

  data () {
    return {
      showStopPipelineModal: false,
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
        },
        {
          label: 'Actions'
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
      },
      pipeline: null
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
            this.pipeline = pipeline.data
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
            this.pipeline = pipeline.data
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

    stopPipeline () {
      this.close()
      this.$http
        .post('/api/v1/pipelinerun/' + this.pipelineID + '/' + this.runID + '/stop', { showProgressBar: false })
        .then(response => {
          if (response.data) {
            this.$router.push({path: '/pipeline/detail', query: { pipelineid: this.pipeline.id, runid: response.data.id }})
          }
        })
        .catch((error) => {
          this.$store.commit('clearIntervals')
          this.$onError(error)
        })
    },

    stopPipelineModal (pipelineID, runID) {
      this.pipelineID = pipelineID
      this.runID = runID
      this.showStopPipelineModal = true
    },

    close () {
      this.showStopPipelineModal = false
      this.$emit('close')
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

      // check if this pipeline has jobs
      if (!jobs) {
        return
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
        // Choose the image for this node and border color
        var nodeImage = require('assets/questionmark.png')
        var borderColor = '#222222'
        if (jobs[i].status) {
          switch (jobs[i].status) {
            case 'success':
              nodeImage = require('assets/success.png')
              break
            case 'failed':
              nodeImage = require('assets/fail.png')
              break
            case 'running':
              nodeImage = require('assets/inprogress.png')
              borderColor = '#e8720b'
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
          },
          color: {
            border: borderColor
          }
        }

        // Add node to nodes list
        nodesArray.push(node)

        // Check if this job has dependencies
        let deps = jobs[i].dependson
        if (!deps) {
          continue
        }

        // Iterate all dependent jobs
        for (let depJobID = 0, depsLength = deps.length; depJobID < depsLength; depJobID++) {
          // iterate again all jobs
          for (let jobID = 0, jobsLength = jobs.length; jobID < jobsLength; jobID++) {
            if (jobs[jobID].id === deps[depJobID].id) {
              // create edge
              let edge = {
                from: jobID,
                to: i
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

    checkPipelineArgs () {
      // check if this pipeline has args
      if (this.pipeline.jobs) {
        for (let x = 0, y = this.pipeline.jobs.length; x < y; x++) {
          if (this.pipeline.jobs[x].args && this.pipeline.jobs[x].args.type !== 'vault') {
            // we found args. Redirect user to params view.
            this.$router.push({path: '/pipeline/params', query: { pipelineid: this.pipeline.id }})
            return
          }
        }
      }

      // No args. Just start pipeline.
      this.startPipeline()
    },

    startPipeline () {
      // Send start request
      this.$http
        .post('/api/v1/pipeline/' + this.pipeline.id + '/start')
        .then(response => {
          if (response.data) {
            this.$router.push({path: '/pipeline/detail', query: { pipelineid: this.pipeline.id, runid: response.data.id }})
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
  .stop-pipeline-modal {
    text-align: center;
    background-color: #2a2735;
  }

</style>
