<template>
  <div class="tile is-ancestor">
    <div class="tile is-vertical is-5">
      <div class="tile">
        <div class="tile is-parent is-vertical">
          <article class="tile is-child notification content-article">
            <p class="title title-text">Compile from source</p>
            <div class="content">
              <label class="label">Copy the link of your <strong>git repo</strong> here.</label>
              <p class="control has-icons-left">
                <input class="input is-medium input-bar" v-focus type="text" placeholder="Link to git repo ...">
                <span class="icon is-small is-left">
                  <i class="fa fa-git"></i>
                </span>
              </p>
              <p class="control">
                <a class="button">
                  <span class="icon">
                    <i class="fa fa-github"></i>
                  </span>
                  <span>Add credentials</span>
                </a>
                <a class="button" style="margin-left: 20px;">
                  <span class="icon">
                    <i class="fa fa-github"></i>
                  </span>
                  <span>Add private key</span>
                </a>
              </p>
              <hr>
              <label class="label">Type the name of your pipeline. You can put your pipelines into folders by defining a path. For example <strong>MyFolder/MyAwesomePipeline</strong>.</label>
              <p class="control has-icons-left">
                <input class="input is-medium input-bar" type="text" placeholder="Pipeline name ...">
                <span class="icon is-small is-left">
                  <i class="fa fa-book"></i>
                </span>
              </p>
              <hr>
              <a class="button">
                <span class="icon">
                  <i class="fa fa-plus"></i>
                </span>
                <span>Create Pipeline</span>
              </a>
            </div>
          </article>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
}
</script>

<style lang="scss" scoped>

</style>
