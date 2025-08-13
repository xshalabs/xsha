<?php

/**
 * 检查后端国际化翻译中未使用的键
 * 新思路：解析所有翻译key，然后逐个搜索项目文件检查是否使用
 */

class BackendI18nChecker {
    private $backendPath;
    private $i18nPath;
    private $srcPath;
    private $allKeys = [];
    private $unusedKeys = [];
    
    public function __construct($basePath = '.') {
        $this->backendPath = rtrim($basePath, '/') . '/backend';
        $this->i18nPath = $this->backendPath . '/i18n/locales';
        $this->srcPath = $this->backendPath;
    }
    
    /**
     * 主执行方法
     */
    public function run() {
        echo "🔍 检查后端国际化翻译中未使用的键...\n\n";
        
        // 检查路径是否存在
        if (!is_dir($this->i18nPath)) {
            die("❌ 错误: 翻译文件目录不存在: {$this->i18nPath}\n");
        }
        
        if (!is_dir($this->srcPath)) {
            die("❌ 错误: 后端源码目录不存在: {$this->srcPath}\n");
        }
        
        // 1. 读取所有翻译键
        $this->loadAllTranslationKeys();
        
        // 2. 逐个检查每个键是否在项目中使用
        $this->checkKeyUsage();
        
        // 3. 输出未使用的键
        $this->outputUnusedKeys();
    }
    
    /**
     * 加载所有翻译键
     */
    private function loadAllTranslationKeys() {
        echo "📖 读取翻译文件...\n";
        
        // 使用 en-US.json 作为基准文件
        $baseFile = $this->i18nPath . '/en-US.json';
        
        if (!file_exists($baseFile)) {
            die("❌ 错误: 基准翻译文件不存在: {$baseFile}\n");
        }
        
        echo "   - 读取基准文件: en-US.json\n";
        
        $content = file_get_contents($baseFile);
        $data = json_decode($content, true);
        
        if ($data === null) {
            die("❌ 错误: 无法解析JSON文件 {$baseFile}\n");
        }
        
        $this->extractKeys($data);
        
        echo "✅ 总共找到 " . count($this->allKeys) . " 个翻译键\n\n";
    }
    
    /**
     * 递归提取所有翻译键
     */
    private function extractKeys($data, $prefix = '') {
        foreach ($data as $key => $value) {
            $fullKey = $prefix ? $prefix . '.' . $key : $key;
            
            if (is_array($value)) {
                $this->extractKeys($value, $fullKey);
            } else {
                // 使用键来避免重复
                $this->allKeys[$fullKey] = true;
            }
        }
    }
    
    /**
     * 检查每个翻译键是否在项目中使用
     */
    private function checkKeyUsage() {
        echo "🔍 逐个检查翻译键是否在项目中使用...\n";
        
        $allKeysArray = array_keys($this->allKeys);
        $totalKeys = count($allKeysArray);
        $processedKeys = 0;
        
        echo "   总共需要检查 {$totalKeys} 个翻译键\n\n";
        
        foreach ($allKeysArray as $key) {
            $processedKeys++;
            
            // 显示进度
            if ($processedKeys % 50 == 0 || $processedKeys == $totalKeys) {
                echo "   处理进度: {$processedKeys}/{$totalKeys} (" . round(($processedKeys / $totalKeys) * 100, 1) . "%)\n";
            }
            
            // 检查该key是否在项目中被使用
            if (!$this->isKeyUsedInProject($key)) {
                $this->unusedKeys[] = $key;
            }
        }
        
        echo "\n✅ 检查完成，发现 " . count($this->unusedKeys) . " 个未使用的翻译键\n\n";
    }
    
    /**
     * 检查指定的key是否在项目中被使用
     */
    private function isKeyUsedInProject($key) {
        // 转义特殊字符以用于grep搜索
        $escapedKey = escapeshellarg($key);
        
        // 使用更精确的grep搜索，排除一些误判情况
        // -n: 显示行号，-H: 显示文件名
        $command = "grep -rn --include='*.go' {$escapedKey} " . escapeshellarg($this->srcPath) . " 2>/dev/null";
        
        // 执行搜索命令
        exec($command, $output, $returnCode);
        
        if (empty($output)) {
            return false;
        }
        
        // 过滤结果，排除一些误判情况
        foreach ($output as $line) {
            // 检查是否是有效的使用（不是在注释中）
            if ($this->isValidUsage($line, $key)) {
                return true;
            }
        }
        
        return false;
    }
    
    /**
     * 检查搜索结果是否是有效的翻译键使用
     */
    private function isValidUsage($line, $key) {
        // 移除文件名和行号前缀
        $content = preg_replace('/^[^:]+:\d+:/', '', $line);
        
        // 排除单行注释
        if (preg_match('/^\s*\/\//', $content)) {
            return false;
        }
        
        // 排除多行注释中的内容
        if (preg_match('/\/\*.*' . preg_quote($key, '/') . '.*\*\//', $content)) {
            return false;
        }
        
        // 检查是否是在字符串或翻译函数调用中使用
        $validPatterns = [
            // i18n.T(lang, "key") 或 i18n.T(lang, "key", args...)
            '/i18n\.T\s*\(\s*[^,]+,\s*["\']' . preg_quote($key, '/') . '["\']/m',
            // h.T("key") 或 h.T("key", args...)
            '/\.T\s*\(\s*["\']' . preg_quote($key, '/') . '["\']/m',
            // h.Response(c, statusCode, "messageKey", ...)
            '/\.Response\s*\(\s*[^,]+,\s*[^,]+,\s*["\']' . preg_quote($key, '/') . '["\']/m',
            // h.ErrorResponse(c, statusCode, "errorKey", ...)
            '/\.ErrorResponse\s*\(\s*[^,]+,\s*[^,]+,\s*["\']' . preg_quote($key, '/') . '["\']/m',
            // T(lang, "key", args...) - global T function
            '/\bT\s*\(\s*[^,]+,\s*["\']' . preg_quote($key, '/') . '["\']/m',
            // errors.New("translationKey")
            '/errors\.New\s*\(\s*["\']' . preg_quote($key, '/') . '["\']\s*\)/m',
            // NewI18nError("translationKey")
            '/NewI18nError\s*\(\s*["\']' . preg_quote($key, '/') . '["\']/m',
            // Key: "translationKey" - struct field assignment
            '/Key\s*:\s*["\']' . preg_quote($key, '/') . '["\']/m',
            // 在翻译映射中使用
            '/["\'][^"\']*["\']\s*:\s*["\']' . preg_quote($key, '/') . '["\']/m'
        ];
        
        foreach ($validPatterns as $pattern) {
            if (preg_match($pattern, $content)) {
                return true;
            }
        }
        
        return false;
    }
    
    /**
     * 输出未使用的翻译键
     */
    private function outputUnusedKeys() {
        if (empty($this->unusedKeys)) {
            echo "🎉 太棒了！所有翻译键都被使用了！\n";
            return;
        }
        
        $totalKeys = count(array_keys($this->allKeys));
        $unusedCount = count($this->unusedKeys);
        $usedCount = $totalKeys - $unusedCount;
        
        echo "❌ 发现 {$unusedCount} 个未使用的翻译键：\n\n";
        
        // 按字母顺序排序
        sort($this->unusedKeys);
        
        // 输出完整的key名
        foreach ($this->unusedKeys as $key) {
            echo "   - {$key}\n";
        }
        
        echo "\n📊 统计信息:\n";
        echo "   - 总翻译键数: {$totalKeys}\n";
        echo "   - 已使用键数: {$usedCount}\n";
        echo "   - 未使用键数: {$unusedCount}\n";
        echo "   - 使用率: " . round(($usedCount / $totalKeys) * 100, 2) . "%\n\n";
        
        echo "💡 清理建议:\n";
        echo "   可以考虑删除这些未使用的翻译键以减少文件大小\n";
        echo "   删除前请确认这些键确实不会在动态生成的场景中使用\n";
        echo "   建议在删除前备份翻译文件\n\n";
    }
}

// 执行检查
try {
    $checker = new BackendI18nChecker();
    $checker->run();
} catch (Exception $e) {
    echo "❌ 执行出错: " . $e->getMessage() . "\n";
    exit(1);
}

?>