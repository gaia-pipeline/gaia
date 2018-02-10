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

  mounted () {
    this.fetchData()
  },

  methods: {
    fetchData () {
      // look up url parameters
      var pipelineID = this.$route.query.pipelineid
      if (!pipelineID) {
        return
      }

      // Get all information from this specific pipeline
      this.$http
        .get('/api/v1/pipelines/detail/' + pipelineID)
        .then(response => {
          this.drawPipelineDetail(response.data)
        })
    },

    drawPipelineDetail (pipeline) {
      // Find container
      var container = document.getElementById('pipeline-detail')

      // prepare data object for vis
      var data = {
        nodes: [],
        edges: []
      }

      // Iterate all jobs of the pipeline
      for (let i = 0, l = pipeline.jobs.length; i < l; i++) {
        // Create nodes object
        var node = {
          id: i,
          shape: 'circularImage',
          image: require('assets/questionmark.png'),
          label: pipeline.jobs[i].title
        }

        // Add node to nodes list
        data.nodes.push(node)

        // Iterate all jobs again to find the next highest job priority
        var highestPrio = null
        for (let x = 0, y = pipeline.jobs.length; x < y; x++) {
          if (pipeline.jobs[x].priority > pipeline.jobs[i].priority && (pipeline.jobs[x].priority < highestPrio || !highestPrio)) {
            highestPrio = pipeline.jobs[x].priority
          }
        }

        // Iterate again all jobs to set all edges
        if (highestPrio) {
          for (let x = 0, y = pipeline.jobs.length; x < y; x++) {
            if (pipeline.jobs[x].priority === highestPrio) {
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

      /* eslint-disable no-unused-vars */
      var network = new Vis.Network(container, data, options)
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
