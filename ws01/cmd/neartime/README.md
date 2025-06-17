# intro
模擬 音訊如何傳入並傳到 asura 進行即時輸出逐字稿

# plan
- 先建立一個純 client 打 asura token -> ws/v3 可以完成整個流程的 case
  - 需要包含來源 pcm 每秒切開並 fmt 輸出到 log