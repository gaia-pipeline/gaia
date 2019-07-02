export default {

  StartPipelineWithArgsCheck (context, pipeline) {
    // check if this pipeline has args
    if (pipeline.jobs) {
      for (let pipelineCurr = 0; pipelineCurr < pipeline.jobs.length; pipelineCurr++) {
        if (pipeline.jobs[pipelineCurr].args) {
          for (let argsCurr = 0; argsCurr < pipeline.jobs[pipelineCurr].args.length; argsCurr++) {
            if (pipeline.jobs[pipelineCurr].args[argsCurr].type !== 'vault') {
              // we found args. Redirect user to params view.
              context.$router.push({path: '/pipeline/params', query: {pipelineid: pipeline.id}})
              return
            }
          }
        }
      }
    }

    // Start the pipeline directly.
    this.StartPipeline(context, pipeline)
  },

  StartPipeline (context, pipeline) {
    // Send start request
    context.$http
      .post('/api/v1/pipeline/' + pipeline.id + '/start')
      .then(response => {
        if (response.data) {
          context.$router.push({path: '/pipeline/detail', query: {pipelineid: pipeline.id, runid: response.data.id}})
        }
      })
      .catch((error) => {
        context.$store.commit('clearIntervals')
        context.$onError(error)
      })
  }
}
