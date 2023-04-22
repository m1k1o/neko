export async function getFilesFromDataTansfer(dataTransfer: DataTransfer): Promise<Array<File>> {
  const files: Array<File> = []

  const traverse = (entry: any): Promise<any> => {
    return new Promise((resolve) => {
      if (entry.isFile) {
        entry.file((file: File) => {
          files.push(file)
          resolve(file)
        })
      } else if (entry.isDirectory) {
        const reader = entry.createReader()
        reader.readEntries((entries: any) => {
          const promises = entries.map(traverse)
          Promise.all(promises).then(resolve)
        })
      }
    })
  }

  const promises: Array<Promise<any>> = []
  // Type 'DataTransferItemList' is not an array type or a string type. Use compiler option '--downlevelIteration' to allow iterating of iterators.
  // @ts-ignore
  for (const item of dataTransfer.items) {
    if ('webkitGetAsEntry' in item) {
      promises.push(traverse(item.webkitGetAsEntry()))
    } else if ('getAsEntry' in item) {
      // @ts-ignore
      promises.push(traverse(item.getAsEntry()))
    } else break
  }

  if (promises.length === 0) {
    // Type 'FileList' is not an array type or a string type. Use compiler option '--downlevelIteration' to allow iterating of iterators.
    // @ts-ignore
    return [...dataTransfer.files]
  }

  await Promise.all(promises)
  return files
}
