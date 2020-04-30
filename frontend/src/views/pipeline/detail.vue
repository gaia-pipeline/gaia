<template>
  <div class="tile is-ancestor">
    <div class="tile is-vertical">
      <div class="tile is-parent">
        <a class="button is-primary" @click="checkPipelineArgsAndStartPipeline" style="margin-right: 10px;">
          <span class="icon">
            <i class="fa fa-play-circle"></i>
          </span>
          <span>Start Pipeline</span>
        </a>
        <a class="button is-green-button" @click="jobLog" v-if="runID">
          <span class="icon">
            <i class="fa fa-terminal"></i>
          </span>
          <span>Show Logs (Run :{{runID}})</span>
        </a>
      </div>
       <div class="tile is-parent" v-if="pipeline">
        <article class="tile is-child notification content-article box">
          <table class="pipeline-detail-table">
            <tr><th>Name</th><td>{{pipeline.name}}</td></tr>
            <tr><th>Type</th><td>{{pipeline.type}}</td></tr>
            <tr><th>Repo</th><td>{{pipeline.repo.url}}</td></tr>
            <tr><th>Branch</th><td>{{pipeline.repo.selectedbranch}}</td></tr>
            <tr><th>Created</th><td><span :title="pipeline.created" v-tippy="{ arrow : true,  animation : 'shift-away'}">{{
              convertTime(pipeline.created) }}</span></td></tr>
            <tr><th>Trigger Token</th><td>{{pipeline.trigger_token}}</td></tr>

            <tr v-if="lastSuccessfulRun">
              <th>Last Successful Run</th>
              <td>
                <router-link :to="{ path: '/pipeline/detail', query: { pipelineid: pipelineID, runid: lastSuccessfulRun.id }}"
                             class="is-blue">
                  {{ lastSuccessfulRun.id }}
                </router-link></td>
            </tr>
            <tr v-if="lastRun">
              <th>Last Run</th>
              <td>
                <router-link :to="{ path: '/pipeline/detail', query: { pipelineid: pipelineID, runid: lastRun.id }}"
                             class="is-blue">
                  {{ lastRun.id }}</router-link>
                [<span v-if="lastRun.status === 'success'" style="color: green;">{{ lastRun.status }}</span>
                <span v-else-if="lastRun.status === 'failed'" style="color: red;">{{ lastRun.status }}</span>
                <span v-else-if="lastRun.status === 'cancelled'" style="color: yellow;">{{ lastRun.status }}</span>
                <span v-else>{{ lastRun.status }}</span> ]
              </td>
            </tr>
          </table>
        </article>
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
            :columns="runsColumns"
            :rows="runsRows"
            :pagination-options="{
              enabled: true,
              mode: 'records'
            }"
            :search-options="{enabled: true, placeholder: 'Search ...'}"
            :sort-options="{
              enabled: true,
              initialSortBy: {field: 'id', type: 'desc'}
            }"
            styleClass="table table-grid table-own-bordered">
            <template slot="table-row" slot-scope="props">
              <span v-if="props.column.field === 'id'">
                <router-link :to="{ path: '/pipeline/detail', query: { pipelineid: pipelineID, runid: props.row.id }}"
                             class="is-blue">
                  {{ props.row.id }}
                </router-link>
              </span>
              <span v-if="props.column.field === 'status'">
                <span v-if="props.row.status === 'success'" style="color: green;">{{ props.row.status }}</span>
                <span v-else-if="props.row.status === 'failed'" style="color: red;">{{ props.row.status }}</span>
                <span v-else-if="props.row.status === 'cancelled'" style="color: yellow;">{{ props.row.status }}</span>
                <span v-else>{{ props.row.status }}</span>
              </span>
              <span v-if="props.column.field === 'duration'">{{ calculateDuration(props.row.startdate, props.row.finishdate) }}</span>
              <span v-if="props.column.field === 'reason'">{{ props.row.started_reason }}</span>
              <span v-if="props.column.field === 'action'">
                <a v-on:click="stopPipelineModal(pipelineID, props.row.id)"><i class="fa fa-ban"
                                                                               style="color: whitesmoke;"></i></a>
              </span>
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
import { VueGoodTable } from 'vue-good-table'
import 'vue-good-table/dist/vue-good-table.css'
import moment from 'moment'
import helper from '../../helper'
import VueTippy from 'vue-tippy'

Vue.use(VueTippy)

export default {
  components: {
    Modal,
    VueGoodTable
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
          label: 'Duration',
          field: 'duration'
        },
        {
          label: 'Reason',
          field: 'reason'
        },
        {
          label: 'Action',
          field: 'action'
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
          arrows: { to: true }
        }
      },
      pipeline: null,
      lastSuccessfulRun: null,
      lastRun: null
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
        Promise.all([this.getPipeline(pipelineID), this.getPipelineRun(pipelineID, runID), this.getPipelineRuns(pipelineID)])
          .then(values => {
            // We only redraw the pipeline if pipeline is running
            var pipeline = values[0]
            var pipelineRun = values[1]
            var pipelineRuns = values[2]
            if (pipelineRun.data.status !== 'running' && !this.lastRedraw) {
              this.drawPipelineDetail(pipeline.data, pipelineRun.data)
              this.lastRedraw = true
            } else if (pipelineRun.data.status === 'running') {
              this.lastRedraw = false
              this.drawPipelineDetail(pipeline.data, pipelineRun.data)
            }
            this.runsRows = pipelineRuns.data
            let tempLastSuccessfulRunId = -1
            let tempLastRunId = -1
            for (let runI = 0; runI < pipelineRuns.data.length; runI++) {
              if (pipelineRuns.data[runI].status === 'success') {
                if (pipelineRuns.data[runI].id > tempLastSuccessfulRunId) {
                  this.lastSuccessfulRun = pipelineRuns.data[runI]
                  tempLastSuccessfulRunId = pipelineRuns.data[runI].id
                }
              }

              if (pipelineRuns.data[runI].id > tempLastRunId) {
                this.lastRun = pipelineRuns.data[runI]
                tempLastRunId = pipelineRuns.data[runI].id
              }
            }
            this.pipeline = pipeline.data
          })
          .catch((error) => {
            this.$store.commit('clearIntervals')
            this.$onError(error)
          })
      } else {
        // Do concurrent request
        Promise.all([this.getPipeline(pipelineID), this.getPipelineRuns(pipelineID)])
          .then(values => {
            var pipeline = values[0]
            var pipelineRuns = values[1]
            if (!this.lastRedraw) {
              this.drawPipelineDetail(pipeline.data, null)
              this.lastRedraw = true
            }

            // Are runs available?
            if (pipelineRuns.data) {
              this.runsRows = pipelineRuns.data
              let tempLastSuccessfulRunId = -1
              let tempLastRunId = -1
              for (let runI = 0; runI < pipelineRuns.data.length; runI++) {
                if (pipelineRuns.data[runI].status === 'success') {
                  if (pipelineRuns.data[runI].id > tempLastSuccessfulRunId) {
                    this.lastSuccessfulRun = pipelineRuns.data[runI]
                    tempLastSuccessfulRunId = pipelineRuns.data[runI].id
                  }
                }

                if (pipelineRuns.data[runI].id > tempLastRunId) {
                  this.lastRun = pipelineRuns.data[runI]
                  tempLastRunId = pipelineRuns.data[runI].id
                }
              }
            }
            this.pipeline = pipeline.data
          })
          .catch((error) => {
            this.$store.commit('clearIntervals')
            this.$onError(error)
          })
      }
    },

    getPipeline (pipelineID) {
      return this.$http.get('/api/v1/pipeline/' + pipelineID, { params: { hideProgressBar: true } })
    },

    getPipelineRun (pipelineID, runID) {
      return this.$http.get('/api/v1/pipelinerun/' + pipelineID + '/' + runID, { params: { hideProgressBar: true } })
    },

    stopPipeline () {
      this.close()
      this.$http
        .post('/api/v1/pipelinerun/' + this.pipelineID + '/' + this.runID + '/stop', { params: { hideProgressBar: true } })
        .then(response => {
          if (response.data) {
            this.$router.push({
              path: '/pipeline/detail',
              query: { pipelineid: this.pipeline.id, runid: response.data.id }
            })
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
      return this.$http.get('/api/v1/pipelinerun/' + pipelineID, { params: { hideProgressBar: true } })
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
        var nodeImage = require('../../assets/images/questionmark.png')
        var borderColor = '#222222'
        if (jobs[i].status) {
          switch (jobs[i].status) {
            case 'success':
              nodeImage = require('../../assets/images/success.png')
              break
            case 'failed':
              nodeImage = require('../../assets/images/fail.png')
              break
            case 'running':
              nodeImage = require('../../assets/images/inprogress.png')
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
      if (moment(startdate).valueOf() < 0) {
        startdate = moment()
      }
      if (moment(finishdate).valueOf() < 0) {
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
      this.$router.push({ path: '/pipeline/log', query: { pipelineid: this.pipelineID, runid: this.runID } })
    },

    checkPipelineArgsAndStartPipeline () {
      helper.StartPipelineWithArgsCheck(this, this.pipeline)
    },

    convertTime (time) {
      return moment(time).fromNow()
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

  .pipeline-detail-table {
    width: 100%;
    table-layout: auto;
    border: 1px solid #000000;
    background-color: #19191B;
    border-radius: 6px;
    border-collapse: separate !important;
    th {
      color: #4da2fc;
    }
    td, th {
      padding: 10px;
      padding-left: 15px;
      border-bottom: 1px solid #000000;
    }
  }
</style>
