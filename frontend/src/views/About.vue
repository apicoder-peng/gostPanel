<template>
  <div class="about-container">
    <el-card class="about-card">
      <template #header>
        <div class="card-header">
          <span>关于系统</span>
        </div>
      </template>
      
      <el-tabs v-model="activeTab">
        <el-tab-pane label="项目介绍" name="info">
          <div class="content-section">
            <h3>Gost Panel</h3>
            <p>现代化 Gost v3 端口转发管理面板</p>
            
            <h4>🚀 核心特性</h4>
            <ul>
              <li>⚡ <strong>极简管理</strong> - 统一管理多个 Gost 节点及其转发规则。</li>
              <li>🔗 <strong>多跳隧道</strong> - 支持复杂的出口/中转隧道配置（入口 -> 出口 -> 目标）。</li>
              <li>📊 <strong>实时延迟</strong> - 每 5 秒自动检测节点与转发链路的 Ping 值。</li>
              <li>🛡️ <strong>安全可靠</strong> - 支持 API 认证，内置操作审计日志。</li>
              <li>🐳 <strong>快速部署</strong> - 全量支持 Docker 一键运行及 GitHub Action 自动构建。</li>
            </ul>
            
            <p>Github: <a href="https://github.com/qiuapeng921/gostPanel" target="_blank" class="link">https://github.com/qiuapeng921/gostPanel</a></p>
          </div>
        </el-tab-pane>
        
        <el-tab-pane label="开源协议" name="license">
          <div class="license-content">
            <pre>{{ licenseText }}</pre>
          </div>
        </el-tab-pane>
        
        <el-tab-pane label="法律声明" name="legal">
          <div class="legal-content">
            <el-alert
              title="免责声明"
              type="warning"
              :closable="false"
              show-icon
            >
              <template #default>
                <div class="legal-text">
                  <p>本项目（Gost Panel）仅供计算机网络技术学习、科研以及合法的网络管理使用。</p>
                  <p>1. 用户在使用本软件时，必须严格遵守当地法律法规。</p>
                  <p>2. 严禁使用本软件进行任何非法活动，包括但不限于非法入侵、网络攻击、绕过网络监管等。</p>
                  <p>3. 作者不对用户使用本软件产生的任何后果负责。</p>
                  <p>4. 下载、安装或使用本软件即视为您已同意本免责声明。</p>
                </div>
              </template>
            </el-alert>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const activeTab = ref('info')

const licenseText = `MIT License

Copyright (c) 2026 qiuapeng921

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.`
</script>

<style scoped>
.about-container {
  padding: 0;
  height: 100%;
}

.about-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

:deep(.el-card__body) {
  flex: 1;
  overflow: hidden;
}

.content-section {
  line-height: 1.8;
  color: #606266;
  padding: 10px;
}

.content-section h3 {
  color: #303133;
  margin-bottom: 15px;
  font-size: 24px;
}

.content-section h4 {
  color: #303133;
  margin: 24px 0 16px;
  font-size: 18px;
}

.content-section ul {
  padding-left: 20px;
  margin-bottom: 20px;
}

.content-section li {
  margin-bottom: 12px;
}

.version-info {
  margin-top: 40px;
  padding-top: 20px;
  border-top: 1px solid #EBEEF5;
  color: #909399;
  font-size: 14px;
}

.link {
  color: #409eff;
  text-decoration: none;
}

.link:hover {
  text-decoration: underline;
}

.license-content {
  height: 100%;
  overflow: auto;
}

.license-content pre {
  background-color: #f5f7fa;
  padding: 20px;
  border-radius: 4px;
  white-space: pre-wrap;
  font-family: 'Courier New', Courier, monospace;
  color: #606266;
  font-size: 14px;
  line-height: 1.6;
  margin: 0;
}

.legal-content {
  padding: 10px;
}

.legal-text {
  margin-top: 10px;
}

.legal-text p {
  margin: 8px 0;
  line-height: 1.6;
}
</style>
