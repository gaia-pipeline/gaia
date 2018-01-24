<template>
  <div class="tile is-ancestor">
    <div class="tile is-vertical is-parent is-5">
      <article class="tile is-child notification content-article">
        <div class="content">
          <label class="label">Copy the link of your <strong>git repo</strong> here.</label>
          <p class="control has-icons-left" v-bind:class="{ 'has-icons-right': gitSuccess }">
            <input class="input is-medium input-bar" v-focus v-model.lazy="gitURL" type="text" placeholder="Link to git repo ...">
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
              <select v-model="gitBranchSelected">
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
          <label class="label">Type the name of your pipeline. You can put your pipelines into folders by defining a path. For example <strong>MyFolder/MyAwesomePipeline</strong>.</label>
          <p class="control has-icons-left">
            <input class="input is-medium input-bar" v-model="pipelineName" type="text" placeholder="Pipeline name ...">
            <span class="icon is-small is-left">
              <i class="fa fa-book"></i>
            </span>
          </p>
          <hr class="dotted-line">
          <a class="button is-primary" v-on:click="createPipeline" v-bind:class="{ 'is-loading': createPipelineStatus }">
            <span class="icon">
              <i class="fa fa-plus"></i>
            </span>
            <span>Create Pipeline</span>
          </a>
        </div>
      </article>

      <div class="tile is-child">
        <article class="tile is-child notification content-article">
          <div class="content">
            Current Status: Some-Status-Here            
            <progress-bar :type="'info'" :size="'medium'" :value="45" :max="100" :show-label="true"></progress-bar>
          </div>
        </article>
      </div>
    </div>

    <div class="tile is-vertical is-parent is-3">
      <div class="tile is-child">
        <article class="tile is-child notification content-article">
          <div class="content">
            Plugin Types here *TODO*
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
                  <input class="input is-medium input-bar" v-focus type="text" v-model="gitUsername" placeholder="Username">
                  <span class="icon is-small is-left">
                    <i class="fa fa-user-circle"></i>
                  </span>
                </p>
                <p class="control has-icons-left">
                  <input class="input is-medium input-bar" type="password" v-model="gitPassword" placeholder="Password">
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
                  <textarea class="textarea input-bar" v-model="privateKey"></textarea>
                </p>
              </div>
              <h2><span>Additional:</span></h2>
              <p class="control has-icons-left" style="padding-bottom: 5px;">
                <input class="input is-medium input-bar" v-focus type="text" v-model="keyUsername" placeholder="Username">
                <span class="icon is-small is-left">
                  <i class="fa fa-user-circle"></i>
                </span>
              </p>
              <p class="control has-icons-left">
                <input class="input is-medium input-bar" type="password" v-model="keyPassword" placeholder="Password">
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

export default {

  data () {
    return {
      gitURL: '',
      gitErrorMsg: '',
      gitSuccess: false,
      gitCredentialsModal: false,
      gitUsername: '',
      gitPassword: '',
      privateKey: '',
      keyUsername: '',
      keyPassword: '',
      gitBranches: [],
      gitBranchSelected: '',
      pipelineName: '',
      createPipelineStatus: '20'
    }
  },

  components: {
    Modal,
    Collapse,
    CollapseItem,
    ProgressBar
  },

  watch: {
    gitURL: function () {
      this.checkGitRepo()
    }
  },

  methods: {
    checkGitRepo () {
      if (this.gitURL === '') {
        return
      }

      // Reset last fetches
      this.gitBranches = []
      this.gitBranchSelected = ''
      this.gitSuccess = false

      var gitrepo = {
        giturl: this.gitURL,
        gituser: this.gitUsername,
        gitpassword: this.gitPassword,
        privatekey: {
          key: this.privateKey,
          username: this.keyUsername,
          password: this.keyPassword
        }
      }

      this.$http.post('/api/v1/pipelines/gitlsremote', gitrepo)
      .then((response) => {
        // Reset error message before
        this.gitErrorMsg = ''
        this.gitSuccess = true

        // Get branches and set to master if available
        this.gitBranches = response.data.gitbranches
        for (var i = 0; i < this.gitBranches.length; i++) {
          if (this.gitBranches[i] === 'refs/heads/master') {
            this.gitBranchSelected = this.gitBranches[i]
          }
        }

        // if we cannot find master
        if (!this.gitBranchSelected && this.gitBranches.length > 0) {
          this.gitBranchSelected = this.gitBranches[0]
        }
      })
      .catch((error) => {
        // Add error message
        this.gitErrorMsg = error.response.data
      })
    },

    createPipeline () {
      var gitrepo = {
        giturl: this.gitURL,
        gituser: this.gitUsername,
        gitpassword: this.gitPassword,
        selectedbranch: this.gitBranchSelected,
        privatekey: {
          key: this.privateKey,
          username: this.keyUsername,
          password: this.keyPassword
        }
      }

      var pipeline = {
        pipelinename: this.pipelineName,
        gitrepo: gitrepo
      }

      this.createPipelineStatus = '20'
      this.$http.post('/api/v1/pipelines/create', pipeline)
      .then((response) => {
        console.log('Pipeline successful created!')
      })
      .catch((error) => {
        console.log(error.response.data)
      })
      this.createPipelineStatus = '100'
    },

    close () {
      this.checkGitRepo()
      this.gitCredentialsModal = false
      this.$emit('close')
    },

    cancel () {
      // cancel means reset all stuff
      this.gitUsername = ''
      this.gitPassword = ''
      this.privateKey = ''
      this.keyUsername = ''
      this.keyPassword = ''

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
  background-image: linear-gradient(to right, black 33%, rgba(255,255,255,0) 0%);
  background-position: bottom;
  background-size: 3px 1px;
  background-repeat: repeat-x;
}

h2 { 
  width:100%; 
  text-align:center; 
  border-bottom: 1px solid #4da2fc; 
  line-height:0.1em; 
  padding-top: 15px;
  margin:10px 0 20px; 
} 

h2 span { 
  background:black;
  color: whitesmoke; 
  padding:0 10px; 
}

.modal-footer {
  height: 35px;
  padding-top: 15px;
}

</style>
