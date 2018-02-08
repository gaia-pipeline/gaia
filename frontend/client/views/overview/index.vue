<template>
  <div class="tile is-ancestor">
    <div class="tile is-vertical">
      <div class="tile">
        <div class="tile is-parent is-3">
          <div class="tile notification content-article">
            <div class="status-display-success"></div>
            <div class="outer-box">
              <div class="outer-box-icon-image">
                <img src="~assets/golang.png" class="outer-box-image">
              </div>
              <div class="name-link">
                <a href="/" class="subtitle">Testpipeline</a>
              </div>
              <div>
                <hr style="color: lightgrey;">
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

.outer-box-icon-image {
  float: left;
  width: 40px;
  height: 40px;
  overflow: hidden;
  border-radius: 50%;
  position: relative;
  border-color: whitesmoke;
  border-style: solid;
  margin-right: 10px;
}

.outer-box-image {
  position: absolute;
  width: 50px;
  height: 50px;
  top: 70%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.name-link {
  margin-top: 5px;
}

</style>
