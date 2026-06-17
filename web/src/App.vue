<template>
  <router-view />
</template>

<script setup>
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

#app {
  width: 100%;
  height: 100vh;
}

/* 移动端全局优化 */
@media (max-width: 768px) {
  /* 表格自适应滚动 */
  .el-table {
    overflow-x: auto;
  }

  .el-table__body-wrapper {
    overflow-x: auto;
  }

  /* 卡片内边距调整 */
  .el-card__body {
    padding: 15px;
  }

  /* 对话框优化 */
  .el-dialog {
    margin: 0 !important;
    max-height: 90vh;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  .el-dialog__body {
    overflow-y: auto;
    flex: 1;
  }

  /* 分页器优化 */
  .el-pagination {
    flex-wrap: wrap;
    justify-content: center !important;
  }

  .el-pagination .el-pagination__sizes {
    margin: 0 5px;
  }

  /* 按钮组优化 */
  .el-button-group {
    display: flex;
    flex-wrap: wrap;
  }

  /* 表单优化 */
  .el-form-item {
    margin-bottom: 16px;
  }

  /* 输入框优化 */
  .el-input__inner,
  .el-textarea__inner {
    font-size: 16px !important; /* 防止 iOS 自动放大 */
  }

  /* 下拉菜单优化 */
  .el-select-dropdown {
    max-width: calc(100vw - 32px);
  }

  /* 消息提示优化 */
  .el-message {
    min-width: auto;
    max-width: calc(100vw - 32px);
  }

  /* 通知优化 */
  .el-notification {
    width: calc(100vw - 32px);
    max-width: 400px;
  }
}

/* 横屏优化 */
@media (max-width: 768px) and (orientation: landscape) {
  .el-dialog {
    max-height: 85vh;
  }
}
</style>
