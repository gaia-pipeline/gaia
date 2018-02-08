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
      nodes: new Vis.DataSet([
        {id: 1, shape: 'circularImage', image: require('assets/success.png'), label: 'Create User'},
        {id: 2, shape: 'circularImage', image: require('assets/success.png'), label: 'Insert Dump'},
        {id: 3, shape: 'circularImage', image: require('assets/fail.png'), label: 'Create Namespace'},
        {id: 4, shape: 'circularImage', image: require('assets/time.png'), label: 'Create Deployment'},
        {id: 5, shape: 'circularImage', image: require('assets/time.png'), label: 'Create Service'},
        {id: 6, shape: 'circularImage', image: require('assets/time.png'), label: 'Create Ingress'},
        {id: 7, shape: 'circularImage', image: require('assets/time.png'), label: 'Clean up'}
      ]),
      edges: new Vis.DataSet([
        {from: 1, to: 2},
        {from: 2, to: 3},
        {from: 3, to: 4},
        {from: 3, to: 5},
        {from: 3, to: 6},
        {from: 4, to: 7},
        {from: 5, to: 7},
        {from: 6, to: 7}
      ])
    }
  },

  mounted () {
    this.fetchData()
  },

  methods: {
    fetchData () {
      // Find container and set data
      var container = document.getElementById('pipeline-detail')
      var data = {
        nodes: this.nodes,
        edges: this.edges
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
