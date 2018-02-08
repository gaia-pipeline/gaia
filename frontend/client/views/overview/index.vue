<template>
  <div class="tile is-ancestor">
    <div class="tile is-vertical">
      <div class="tile">
        <div class="tile is-parent is-3">
          <div class="tile notification content-article">
            <div class="status-display-success"></div>
            <div class="outer-box">
              <div class="outer-box-icon">
                <i class="fa fa-2x fa-battery-full"></i>
              </div>
              <div>
                <a href="/" class="subtitle">Testpipeline</a>
              </div>
              <div style="padding-top: 10px;">
                this is some text ...
              </div>
            </div>
          </div>
        </div>

        <div class="tile is-parent is-3">
          <div class="tile notification content-article">
            <div class="status-display-success"></div>
            <div class="outer-box">
              <div class="outer-box-icon">
                <i class="fa fa-2x fa-battery-full"></i>
              </div>
              <div>
                <a href="/" class="subtitle">Testpipeline</a>
              </div>
              <div style="padding-top: 10px;">
                this is some text ...
              </div>
            </div>
          </div>
        </div>

        <div class="tile is-parent is-3">
          <div class="tile notification content-article">
            <div class="status-display-folder"></div>
            <div class="outer-box">
              <div class="outer-box-icon">
                <i class="fa fa-2x fa-folder"></i>
              </div>
              <div>
                <a href="/" class="subtitle">Testfolder</a>
              </div>
              <div style="padding-top: 10px;">
                this is some text ...
              </div>
            </div>
          </div>
        </div>

        <div class="tile is-parent is-3">
          <div class="tile notification content-article">
            <div class="status-display-folder"></div>
            <div class="outer-box">
              <div class="outer-box-icon">
                <i class="fa fa-2x fa-folder"></i>
              </div>
              <div>
                <a href="/" class="subtitle">Testfolder</a>
              </div>
              <div style="padding-top: 10px;">
                this is some text ...
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="tile">
        <div class="tile is-parent is-3">
          <div class="tile notification content-article">
            <div class="status-display-fail"></div>
            <div class="outer-box">
              <div class="outer-box-icon">
                <i class="fa fa-2x fa-battery-full"></i>
              </div>
              <div>
                <a href="/" class="subtitle">Testpipeline</a>
              </div>
              <div style="padding-top: 10px;">
                this is some text ...
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {

}
</script>

<style lang="scss">

@mixin status-display {
  position: fixed;
  min-width: 50px;
  height: 100%;
  margin-left: -23px;
  margin-top: -20px;
  margin-bottom: -20px;
  border-top-left-radius: 3px;
  border-bottom-left-radius: 3px;
  margin-right: 10px;
}

.status-display-success {
  @include status-display();
  background-color: rgb(49, 196, 49);  
}

.status-display-folder {
  @include status-display();
  background-color: #4da2fc;
}

.status-display-fail {
  @include status-display();
  background-color: #ca280b;
}

.outer-box {
  padding-left: 40px; 
  min-height: 170px;
  width: 100%;
}

.outer-box-icon {
  width: 50px;
  float: left;
}

</style>
