<template>
  <div class="tile is-ancestor">
    <div class="tile is-vertical">
      <div class="tile">
        <div class="tile is-vertical is-parent is-12">
          <article class="tile is-child notification content-article">
            <div id="pipeline-detail"></div>
          </article>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import Vis from 'vis'

export default {

  data () {
    return {
      pipelineView: null
    }
  },

  mounted () {
    this.fetchData()

    // periodically update view
    setInterval(function () {
      this.fetchData()
    }.bind(this), 3000)
  },

  methods: {
    fetchData () {
      // look up url parameters
      var pipelineID = this.$route.query.pipelineid
      if (!pipelineID) {
        return
      }

      // runID is optional
      var runID = this.$route.query.runid

      // Get all information from this specific pipeline
      var pipeline = null
      this.$http
        .get('/api/v1/pipelines/detail/' + pipelineID)
        .then(response => {
          this.pipeline = response.data
        })

      // If runid was set, look up this run
      var pipelineRun = null
      if (runID) {
        this.$http
          .get('/api/v1/pipelines/detail/' + pipelineID + '/' + runID)
          .then(response => {
            this.pipelineRun = response.data
          })
      }

      // Draw pipeline view
      this.drawPipelineDetail(pipeline, pipelineRun)
    },

    drawPipelineDetail (pipeline, pipelineRun) {
      // Check if pipelineRun was set
      var jobs = null
      if (pipelineRun) {
        jobs = pipelineRun.jobs
      } else {
        jobs = pipeline.jobs
      }

      // prepare data object for vis
      var data = {
        nodes: [],
        edges: []
      }

      // Iterate all jobs of the pipeline
      for (let i = 0, l = jobs.length; i < l; i++) {
        // Choose the image for this node
        var nodeImage = require('assets/questionmark.png')
        if (jobs[i].status) {
          switch (jobs[i].status) {
            case 'success': 
              nodeImage = require('assets/success.png')
            case 'failed':
              nodeImage = require('assets/failed.png')
          }
        }

        // Create nodes object
        var node = {
          id: i,
          shape: 'circularImage',
          image: nodeImage,
          label: jobs[i].title
        }

        // Add node to nodes list
        data.nodes.push(node)

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
              var edge = {
                from: i,
                to: x
              }

              // add edge to edges list
              data.edges.push(edge)
            }
          }
        }
      }

      // Define vis options
      var options = {
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

      // If pipelineView already exist, just update it
      if (this.pipelineView) {
        // Redraw
        this.pipelineView.Redraw()
      } else {
        // Find container
        var container = document.getElementById('pipeline-detail')
        this.pipelineView = new Vis.Network(container, data, options)
      }
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
