<template>
  <div class="tile is-ancestor">
    <div class="tile is-vertical">
      <div class="tile">
        <div class="tile is-vertical is-parent is-5">
          <article class="tile is-child notification content-article">
            <div class="content">
              <label class="label">Copy the link of your
                <strong>git repo</strong> here.</label>
              <p class="control has-icons-left" v-bind:class="{ 'has-icons-right': gitSuccess }">
                <input class="input is-medium input-bar" v-focus v-model.lazy="giturl" type="text" placeholder="Link to git repo ...">
                <span class="icon is-small is-left">
                  <i class="fa fa-git"></i>
                </span>
                <span v-if="gitSuccess" class="icon is-small is-right is-blue">
                  <i class="fa fa-check"></i>
                </span>
              </p>
              <span style="color: red" v-if="gitErrorMsg">Cannot access git repo: {{ gitErrorMsg }}</span>
              <div v-if="gitBranches.length > 0">
                <span>Branch:</span>
                <div class="select is-fullwidth">
                  <select v-model="pipeline.gitrepo.selectedbranch">
                    <option v-for="branch in gitBranches" :key="branch" :value="branch">{{ branch }}</option>
                  </select>
                </div>
              </div>
              <p class="control" style="padding-top: 10px;">
                <a class="button is-primary" v-on:click="showCredentialsModal">
                  <span class="icon">
                    <i class="fa fa-certificate"></i>
                  </span>
                  <span>Add credentials</span>
                </a>
              </p>
              <hr class="dotted-line">
              <label class="label">Type the name of your pipeline. You can put your pipelines into folders by defining a path. For example
                <strong>MyFolder/MyAwesomePipeline</strong>.</label>
              <p class="control has-icons-left" v-bind:class="{ 'has-icons-right': pipelineNameSuccess }">
                <input class="input is-medium input-bar" v-model.lazy="pipelinename" type="text" placeholder="Pipeline name ...">
                <span class="icon is-small is-left">
                  <i class="fa fa-book"></i>
                </span>
                <span v-if="pipelineNameSuccess" class="icon is-small is-right is-blue">
                  <i class="fa fa-check"></i>
                </span>
              </p>
              <span style="color: red" v-if="pipelineErrorMsg">Pipeline Name incorrect: {{ pipelineErrorMsg }}</span>
              <hr class="dotted-line">
              <a class="button is-primary" v-on:click="createPipeline" v-bind:class="{ 'is-disabled': !gitSuccess || !pipelineNameSuccess }">
                <span class="icon">
                  <i class="fa fa-plus"></i>
                </span>
                <span>Create Pipeline</span>
              </a>
            </div>
          </article>
        </div>

        <div class="tile is-parent is-3">
          <article class="tile is-child notification content-article box">
            <p class="subtitle">Select pipeline language</p>
            <div class="content" style="display: flex;">
              <div class="pipelinetype tippy" title="Golang" v-on:click="pipeline.pipelinetype = 'golang'" v-bind:class="{ pipelinetypeactive: pipeline.pipelinetype === 'golang' }" data-tippy-hideOnClick="false">
                <img src="~assets/golang.png" class="typeimage">
              </div>
              <div class="pipelinetype tippy" title="Python (not yet supported)" v-bind:class="{ pipelinetypeactive: pipeline.pipelinetype === 'python' }" data-tippy-hideOnClick="false">
                <img src="~assets/python.png" class="typeimage typeimagenotyetsupported">
              </div>
              <div class="pipelinetype tippy" title="Java (not yet supported)" v-bind:class="{ pipelinetypeactive: pipeline.pipelinetype === 'java' }" data-tippy-hideOnClick="false">
                <img src="~assets/java.png" class="typeimage typeimagenotyetsupported">
              </div>
            </div>
            <div class="content" style="display: flex;">
              <div class="pipelinetype tippy" title="C++ (not yet supported)" v-bind:class="{ pipelinetypeactive: pipeline.pipelinetype === 'cplusplus' }" data-tippy-hideOnClick="false">
                <img src="~assets/cplusplus.png" class="typeimage typeimagenotyetsupported">
              </div>
              <div class="pipelinetype tippy" title="Node.js (not yet supported)" v-bind:class="{ pipelinetypeactive: pipeline.pipelinetype === 'nodejs' }" data-tippy-hideOnClick="false">
                <img src="~assets/nodejs.png" class="typeimage typeimagenotyetsupported">
              </div>
            </div>
          </article>
        </div>
      </div>

      <div class="tile is-parent is-8">
        <article class="tile is-child notification content-article box">
          <p class="subtitle">Pipelines history</p>
          <div class="content">
            <div class="table-responsive">
              <table class="table">
                <thead>
                  <tr>
                    <th class="th">Name</th>
                    <th class="th">Status</th>
                    <th class="th">Type</th>
                    <th class="th">Creation date</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-bind:class="{ blink: pipeline.status < 100 }" v-for="(pipeline, index) in createdPipelines" :key="index">
                    <td v-bind:class="{ th: pipeline.status === 100 }">
                      {{ pipeline.pipelinename }}
                    </td>
                    <td class="th">
                      <div v-if="pipeline.status < 100">
                        <progress-bar :type="'info'" :size="'small'" :value="pipeline.status" :max="100" :show-label="false"></progress-bar>
                      </div>
                      <div v-if="pipeline.status === 100">
                        <span>Completed</span>
                      </div>
                    </td>
                    <td v-bind:class="{ th: pipeline.status === 100 }">
                      {{ pipeline.pipelinetype }}
                    </td>
                    <td v-bind:class="{ th: pipeline.status === 100 }">
                      {{ pipeline.creationdate }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </article>
      </div>
    </div>

    <!-- Credentials modal -->
    <modal :visible="gitCredentialsModal" class="modal-z-index" @close="close">
      <div class="box credentials-modal">
        <div class="block credentials-modal-content">
          <collapse accordion is-fullwidth>
            <collapse-item title="Basic Authentication" selected>
              <div class="credentials-modal-content">
                <label class="label" style="text-align: left;">Add credentials for basic authentication:</label>
                <p class="control has-icons-left" style="padding-bottom: 5px;">
                  <input class="input is-medium input-bar" v-focus type="text" v-model="pipeline.gitrepo.gituser" placeholder="Username">
                  <span class="icon is-small is-left">
                    <i class="fa fa-user-circle"></i>
                  </span>
                </p>
                <p class="control has-icons-left">
                  <input class="input is-medium input-bar" type="password" v-model="pipeline.gitrepo.gitpassword" placeholder="Password">
                  <span class="icon is-small is-left">
                    <i class="fa fa-lock"></i>
                  </span>
                </p>
              </div>
            </collapse-item>
            <collapse-item title="SSH Key">
              <label class="label" style="text-align: left;">Instead of using basic authentication, provide a pem encoded private key.</label>
              <div class="block credentials-modal-content">
                <p class="control">
                  <textarea class="textarea input-bar" v-model="pipeline.gitrepo.privatekey.key"></textarea>
                </p>
              </div>
              <h2>
                <span>Additional:</span>
              </h2>
              <p class="control has-icons-left" style="padding-bottom: 5px;">
                <input class="input is-medium input-bar" v-focus type="text" v-model="pipeline.gitrepo.privatekey.username" placeholder="Username">
                <span class="icon is-small is-left">
                  <i class="fa fa-user-circle"></i>
                </span>
              </p>
              <p class="control has-icons-left">
                <input class="input is-medium input-bar" type="password" v-model="pipeline.gitrepo.privatekey.password" placeholder="Password">
                <span class="icon is-small is-left">
                  <i class="fa fa-lock"></i>
                </span>
              </p>
            </collapse-item>
          </collapse>
          <div class="modal-footer">
            <div style="float: left;">
              <button class="button is-primary" v-on:click="close">Add Credentials</button>
            </div>
            <div style="float: right;">
              <button class="button is-danger" v-on:click="cancel">Cancel</button>
            </div>
          </div>
        </div>
      </div>
    </modal>
  </div>
</template>

<script>
import { Modal } from 'vue-bulma-modal'
import { Collapse, Item as CollapseItem } from 'vue-bulma-collapse'
import ProgressBar from 'vue-bulma-progress-bar'
import Tippy from 'tippy.js'

export default {
  data () {
    return {
      gitErrorMsg: '',
      gitSuccess: false,
      gitCredentialsModal: false,
      gitBranches: [],
      giturl: '',
      pipelinename: '',
      pipelineNameSuccess: false,
      pipelineErrorMsg: '',
      createdPipelines: [],
      pipeline: {
        pipelinename: '',
        pipelinetype: 'golang',
        gitrepo: {
          giturl: '',
          gituser: '',
          gitpassword: '',
          selectedbranch: '',
          privatekey: {
            key: '',
            username: '',
            password: ''
          }
        }
      }
    }
  },

  components: {
    Modal,
    Collapse,
    CollapseItem,
    ProgressBar,
    Tippy
  },

  mounted () {
    // tippy
    Tippy('.tippy', {
      placement: 'top',
      animation: 'scale',
      duration: 500,
      arrow: true
    })

    // created pipelines history
    this.fetchData()
  },

  watch: {
    giturl: function () {
      this.checkGitRepo()
    },
    pipelinename: function () {
      this.checkPipelineNameAvailable()
    },
    '$route': 'fetchData'
  },

  methods: {
    fetchData () {
      this.$http
        .get('/api/v1/pipelines/created')
        .then(response => {
          this.createdPipelines = response.data
        })
        .catch(error => {
          console.log(error.response.data)
        })
    },

    checkGitRepo () {
      if (this.giturl === '') {
        return
      }

      // copy giturl into our struct
      this.pipeline.gitrepo.giturl = this.giturl

      // Reset last fetches
      this.gitBranches = []
      this.pipeline.gitrepo.selectedbranch = ''
      this.gitSuccess = false

      this.$http
        .post('/api/v1/pipelines/gitlsremote', this.pipeline.gitrepo)
        .then(response => {
          // Reset error message before
          this.gitErrorMsg = ''
          this.gitSuccess = true

          // Get branches and set to master if available
          this.gitBranches = response.data
          for (var i = 0; i < this.gitBranches.length; i++) {
            if (this.gitBranches[i] === 'refs/heads/master') {
              this.pipeline.gitrepo.selectedbranch = this.gitBranches[i]
            }
          }

          // if we cannot find master
          if (
            !this.pipeline.gitrepo.selectedbranch &&
            this.gitBranches.length > 0
          ) {
            this.pipeline.gitrepo.selectedbranch = this.gitBranches[0]
          }
        })
        .catch(error => {
          // Add error message
          this.gitErrorMsg = error.response.data
        })
    },

    checkPipelineNameAvailable () {
      // copy pipeline name into struct
      this.pipeline.pipelinename = this.pipelinename

      // Request for availability
      this.$http
        .post('/api/v1/pipelines/name', this.pipeline)
        .then(response => {
          // pipeline name valid and available
          this.pipelineErrorMsg = ''
          this.pipelineNameSuccess = true
        })
        .catch(error => {
          this.pipelineErrorMsg = error.response.data
          this.pipelineNameSuccess = false
        })
    },

    createPipeline () {
      // copy giturl into our struct and pipeline name
      this.pipeline.gitrepo.giturl = this.giturl
      this.pipeline.pipelinename = this.pipelinename

      // Start the create pipeline process in the backend
      this.$http
        .post('/api/v1/pipelines/create', this.pipeline)
        .then(response => {
          // Run fetchData to see the pipeline in our history table
          this.fetchData()
        })
        .catch(error => {
          console.log(error.response.data)
        })
    },

    close () {
      this.checkGitRepo()
      this.gitCredentialsModal = false
      this.$emit('close')
    },

    cancel () {
      // cancel means reset all stuff
      this.pipeline.gitrepo.gituser = ''
      this.pipeline.gitrepo.gitpassword = ''
      this.pipeline.gitrepo.privatekey.key = ''
      this.pipeline.gitrepo.privatekey.username = ''
      this.pipeline.gitrepo.privatekey.password = ''

      this.close()
    },

    showCredentialsModal () {
      this.gitCredentialsModal = true
    }
  }
}
</script>

<style lang="scss" scoped>
.credentials-modal {
  text-align: center;
  background-color: #2a2735;
}

.credentials-modal-content {
  margin: auto;
  padding: 10px;
}

.dotted-line {
  background-image: linear-gradient(
    to right,
    black 33%,
    rgba(255, 255, 255, 0) 0%
  );
  background-position: bottom;
  background-size: 3px 1px;
  background-repeat: repeat-x;
}

h2 {
  width: 100%;
  text-align: center;
  border-bottom: 1px solid #4da2fc;
  line-height: 0.1em;
  padding-top: 15px;
  margin: 10px 0 20px;
}

h2 span {
  background: black;
  color: whitesmoke;
  padding: 0 10px;
}

.modal-footer {
  height: 35px;
  padding-top: 15px;
}

.pipelinetype {
  height: 100px;
  width: 100px;
  border: 1px solid;
  color: black;
  margin: 0 5px;
  box-shadow: 4px 4px 4px 4px rgba(0, 0, 0, 0.2),
    0 1px 1px 0 rgba(0, 0, 0, 0.19);
}

.pipelinetype:hover {
  -moz-transform: scale(1.1);
  -webkit-transform: scale(1.1);
  transform: scale(1.1);
}

.pipelinetypeactive {
  color: #4da2fc;
  border: 2px solid;
}

.typeimage {
  display: block;
  margin: 0 auto;
  height: 100%;
}

.typeimagenotyetsupported {
  opacity: 0.5;
}

.table {
  background-color: #3f3d49;
  color: #4da2fc;
}

.th {
  border-color: gray;
  color: whitesmoke;
  vertical-align: middle;
}

.content table tr:hover {
  background-color: #2a2735;
}

.progress-container {
  margin-bottom: 0px;
}

.blink {
  animation: blink 700ms infinite alternate;
}

@keyframes blink {
  from {
    opacity: 1;
  }
  to {
    opacity: 0.2;
  }
}
</style>
