export default {

  StartPipelineWithArgsCheck (context, pipeline) {
    // check if this pipeline has args
    if (pipeline.jobs) {
      for (let pipelineCurr = 0; pipelineCurr < pipeline.jobs.length; pipelineCurr++) {
        if (pipeline.jobs[pipelineCurr].args) {
          for (let argsCurr = 0; argsCurr < pipeline.jobs[pipelineCurr].args.length; argsCurr++) {
            if (pipeline.jobs[pipelineCurr].args[argsCurr].type !== 'vault') {
              // we found args. Redirect user to params view.
              context.$router.push({ path: '/pipeline/params', query: { pipelineid: pipeline.id, docker: pipeline.docker } })
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
      .post('/api/v1/pipeline/' + pipeline.id + '/start', [{ key: 'docker', value: this.docker ? '1' : '0' }])
      .then(response => {
        if (response.data) {
          context.$router.push({ path: '/pipeline/detail', query: { pipelineid: pipeline.id, runid: response.data.id } })
        }
      })
      .catch((error) => {
        context.$store.commit('clearIntervals')
        context.$onError(error)
      })
  },

  PullPipeline (context, pipeline) {
    // Send pull request
    context.$http
      .post('/api/v1/pipeline/' + pipeline.id + '/pull', { docker: pipeline.docker })
      .then(response => {
        context.$notify({
          title: 'Successfully pulled new code',
          message: `Pipeline "${pipeline.name}" has been updated successfully.`,
          type: 'success'
        })
      })
      .catch((error) => {
        context.$store.commit('clearIntervals')
        context.$onError(error)
      })
  }
}
